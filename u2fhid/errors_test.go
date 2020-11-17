package u2fhid

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateError(t *testing.T) {
	t.Run("error packets are generated correctly", func(t *testing.T) {
		p := initPacket{
			ChannelID: [4]byte{
				1,
				2,
				3,
				4,
			},
			Cmd:           cmdPing,
			PayloadLength: 42,
			Data:          bytes.Repeat([]byte{42}, 42),
		}

		b := generateError(other, p)

		require.Len(t, b, 1)

		pkt := b[0]

		require.Len(t, pkt, 8) // 7 bytes of header + 1 for error

		require.Equal(t, pkt[0:4], p.ChannelID[:]) // channelID
		require.Equal(t, pkt[4], uint8(cmdError))  // command
		require.Equal(t, pkt[5:7], []byte{0, 1})   // packet count
		require.Equal(t, pkt[7], uint8(other))     // error number
	})
}
