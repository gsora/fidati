package u2fhid

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_continuationPacket(t *testing.T) {
	cp := continuationPacket{
		ChannelID: [4]byte{
			1,
			2,
			3,
			4,
		},
		SequenceNumber: 42,
		Data:           bytes.Repeat([]byte("data"), 42),
	}

	tests := []struct {
		name string
		f    func(*testing.T)
	}{
		{
			"ChannelBytes",
			func(t *testing.T) {
				require.Equal(t, [4]byte{1, 2, 3, 4}, cp.ChannelBytes())
			},
		},
		{
			"Command",
			func(t *testing.T) {
				require.Zero(t, cp.Command())
			},
		},
		{
			"Length",
			func(t *testing.T) {
				require.Zero(t, cp.Length())
			},
		},
		{
			"Count",
			func(t *testing.T) {
				require.Equal(t, uint16(cp.SequenceNumber), cp.Count())
			},
		},
		{
			"Channel",
			func(t *testing.T) {
				require.Equal(t, uint32(16909060), cp.Channel())
			},
		},
		{
			"Bytes",
			func(t *testing.T) {
				b := cp.Bytes()
				require.Equal(t, []byte{1, 2, 3, 4}, b[:4])
				require.Equal(t, uint8(42), b[4])
				require.Equal(t, bytes.Repeat([]byte("data"), 42), b[5:])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.f)
	}
}

func Test_parseContinuationPkt(t *testing.T) {
	t.Run("parses correctly", func(t *testing.T) {
		b := []byte{1, 2, 3, 4, 42, 42, 42, 42, 42}
		var cp continuationPacket

		require.NotPanics(t, func() {
			cp = parseContinuationPkt(b)
		})

		require.Equal(t, [4]byte{1, 2, 3, 4}, cp.ChannelID)
		require.Equal(t, uint8(42), cp.SequenceNumber)
		require.Equal(t, []byte{42, 42, 42, 42}, cp.Data)
	})
}

func Test_initPacket(t *testing.T) {
	cp := initPacket{
		ChannelID: [4]byte{
			1,
			2,
			3,
			4,
		},
		Cmd:           cmdInit,
		PayloadLength: 42,
		Data:          bytes.Repeat([]byte("data"), 42),
	}

	tests := []struct {
		name string
		f    func(*testing.T)
	}{
		{
			"ChannelBytes",
			func(t *testing.T) {
				require.Equal(t, [4]byte{1, 2, 3, 4}, cp.ChannelBytes())
			},
		},
		{
			"Command",
			func(t *testing.T) {
				require.Equal(t, uint8(cmdInit), cp.Command())
			},
		},
		{
			"Length",
			func(t *testing.T) {
				require.Equal(t, uint16(42), cp.Length())
			},
		},
		{
			"Count",
			func(t *testing.T) {
				require.Zero(t, cp.Count())
			},
		},
		{
			"Channel",
			func(t *testing.T) {
				require.Equal(t, uint32(16909060), cp.Channel())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.f)
	}
}

func Test_parseInitPkt(t *testing.T) {
	t.Run("parses correctly", func(t *testing.T) {
		b := []byte{1, 2, 3, 4, 11, 11, 11, 42, 42, 42, 42}
		var cp initPacket

		require.NotPanics(t, func() {
			cp = parseInitPkt(b)
		})

		require.Equal(t, [4]byte{1, 2, 3, 4}, cp.ChannelID)
		require.Equal(t, u2fHIDCommand(11), cp.Cmd)
		require.Equal(t, uint16(2827), cp.PayloadLength)
		require.Equal(t, []byte{42, 42, 42, 42}, cp.Data)
	})
}

func Test_isInitPkt(t *testing.T) {
	tests := []struct {
		name      string
		cmd       uint8
		assertion require.BoolAssertionFunc
	}{
		{
			"is init packet",
			0b10000000,
			require.True,
		},
		{
			"is not init packet",
			0b00000001,
			require.False,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.assertion(t, isInitPkt(tt.cmd))
		})
	}
}

func Test_split(t *testing.T) {
	tests := []struct {
		name       string
		msg        []byte
		numPackets int
	}{
		{
			"nil msg",
			nil,
			0,
		},
		{
			"all in one packet",
			bytes.Repeat([]byte{1}, 40),
			1,
		},
		{
			"multiple packets",
			bytes.Repeat([]byte{1}, 180),
			4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := [][]byte{}

			require.NotPanics(t, func() {
				d = split(initPacketDataLen, continuationPacketDataLen, tt.msg)
			})

			require.Equal(t, tt.numPackets, len(d))
		})
	}
}

func Test_genPackets(t *testing.T) {
	cmd, chanID := cmdInit, [4]byte{1, 2, 3, 4}

	tests := []struct {
		name         string
		msg          []byte
		pktAmount    int
		errAssertion require.ErrorAssertionFunc
	}{
		{
			"empty message",
			nil,
			0,
			require.Error,
		},
		{
			"single packet generated",
			bytes.Repeat([]byte{1}, 12),
			1,
			require.NoError,
		},
		{
			"multiple packet generated",
			bytes.Repeat([]byte{1}, 4242),
			72,
			require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := genPackets(tt.msg, cmd, chanID)

			tt.errAssertion(t, err)

			if data == nil {
				return
			}

			require.Equal(t, tt.pktAmount, len(data)) // exclude the init packet

			lastIndex := -1
			for i, pkt := range data {
				if i == 0 {
					// first packet has length
					l := pkt[5:7]
					l16 := binary.BigEndian.Uint16(l)
					require.Equal(t, uint16(len(tt.msg)), l16)
					continue
				}

				li := int(pkt[4])
				if lastIndex == -1 {
					lastIndex = li
					continue
				}

				// each sequence number must be distant exactly one from the last
				difference := li - lastIndex
				require.Equal(t, 1, difference)
				lastIndex = li
			}
		})
	}
}
