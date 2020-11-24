package u2fhid

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_zeroPad(t *testing.T) {
	tests := []struct {
		name       string
		b          []byte
		zeroAmount int
		want       []byte
	}{
		{
			"data must be padded",
			bytes.Repeat([]byte{1}, 42),
			22,
			append(bytes.Repeat([]byte{1}, 42), make([]byte, 22)...),
		},
		{
			"data must not be padded",
			bytes.Repeat([]byte{1}, 64),
			0,
			bytes.Repeat([]byte{1}, 64),
		},
		{
			"if more than 64 bytes are given, only the first 64 are returned",
			append(bytes.Repeat([]byte{1}, 64), bytes.Repeat([]byte{2}, 64)...),
			0,
			bytes.Repeat([]byte{1}, 64),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := zeroPad(tt.b)

			zeros := 0
			if len(tt.b) <= 64 {
				zeros = len(p[len(tt.b):])
			}

			require.Equal(t, tt.zeroAmount, zeros)
			require.Equal(t, tt.want, p)
		})
	}
}

func TestHandler_parseMsg(t *testing.T) {
	tests := []test{
		{
			"more than 64 bytes means error",
			func(t *testing.T) {
				h, err := NewHandler(&fakeToken{})
				require.NoError(t, err)

				b := make([]byte, 65)
				d, err := h.parseMsg(b)
				require.Nil(t, d)
				require.Error(t, err)
				require.Contains(t, err.Error(), "wrong message length")
			},
		},
		{
			"init packet not on broadcast channel returns error",
			func(t *testing.T) {
				h, err := NewHandler(&fakeToken{})
				require.NoError(t, err)

				msg := zeroPad([]byte{1, 2, 3, 4, uint8(cmdInit), 0, 8, 1, 2, 3, 4, 5, 6, 7, 8})

				d, err := h.parseMsg(msg)
				require.Error(t, err)
				require.Nil(t, d)
				require.Contains(t, err.Error(), "not on the broadcast channel")
			},
		},
		{
			"init packet is parsed correctly",
			func(t *testing.T) {
				h, err := NewHandler(&fakeToken{})
				require.NoError(t, err)

				channelInt := binary.BigEndian.Uint32([]byte{255, 255, 255, 255})

				msg := zeroPad([]byte{255, 255, 255, 255, uint8(cmdInit), 0, 8, 1, 2, 3, 4, 5, 6, 7, 8})

				d, err := h.parseMsg(msg)
				require.NoError(t, err)
				require.NotNil(t, d)

				require.False(t, h.state.accumulatingMsgs)
				require.Contains(t, h.state.sessions, channelInt)

				s := h.state.sessions[channelInt]

				require.Equal(t, uint64(8), s.total)
				require.Equal(t, []byte{1, 2, 3, 4, 5, 6, 7, 8}, s.data[:8])
			},
		},
		{
			"continuation packet without a previous init packet returns error",
			func(t *testing.T) {
				h, err := NewHandler(&fakeToken{})
				require.NoError(t, err)

				msg := zeroPad([]byte{1, 2, 3, 4, 0, 8, 1, 2, 3, 4, 5, 6, 7, 8})

				d, err := h.parseMsg(msg)
				require.Error(t, err)
				require.Nil(t, d)
				require.Contains(t, err.Error(), "which was not seen before")
			},
		},
		{
			"init packet and then continuation packet are parsed correctly",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				firstHalf := bytes.Repeat([]byte{1}, 57)
				secondHalf := bytes.Repeat([]byte{1}, 5)

				channel := []byte{1, 2, 3, 4}

				channelInt := binary.BigEndian.Uint32(channel)

				initNoPad := append(channel, uint8(cmdMsg))
				initNoPad = append(initNoPad, []byte{0, 62}...)
				initNoPad = append(initNoPad, firstHalf...)
				init := zeroPad(initNoPad)

				contNoPad := append(channel, 0)
				contNoPad = append(contNoPad, secondHalf...)
				cont := zeroPad(contNoPad)

				d, err := h.parseMsg(init)
				require.NoError(t, err)
				require.Nil(t, d)

				require.True(t, h.state.accumulatingMsgs)
				require.Contains(t, h.state.sessions, channelInt)

				s := h.state.sessions[channelInt]

				require.Equal(t, uint64(57+5), s.total)
				require.Equal(t, cmdMsg, s.command)
				require.Equal(t, firstHalf, s.data)
				require.Equal(t, uint64(5), s.leftToRead)

				// set bogus data for token to return
				token.shouldReturnData = true
				token.data = append(firstHalf, secondHalf...)

				d, err = h.parseMsg(cont)
				require.NoError(t, err)
				require.NotNil(t, d)

				require.False(t, h.state.accumulatingMsgs)
				require.Contains(t, h.state.sessions, channelInt)

				s = h.state.sessions[channelInt]

				require.Equal(t, uint64(57+5), s.total)
				require.Equal(t, cmdMsg, s.command)
				require.Equal(t, append(firstHalf, secondHalf...), s.data)
				require.Equal(t, uint64(0), s.leftToRead)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.f)
	}
}

func TestHandler_handleContinuationPacket(t *testing.T) {
	tests := []test{
		{
			"continuation packet with previously not seen channel ID",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				secondHalf := bytes.Repeat([]byte{1}, 5)

				channel := []byte{1, 2, 3, 4}

				contNoPad := append(channel, 0)
				contNoPad = append(contNoPad, secondHalf...)
				cont := zeroPad(contNoPad)

				d, err := h.handleContinuationPacket(cont)
				require.Error(t, err)
				require.Contains(t, err.Error(), "not seen before")

				require.Nil(t, d)
			},
		},
		{
			"continuation packet with previously seen channel ID has unexpected sequence number",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				firstHalf := bytes.Repeat([]byte{1}, 57)
				secondHalf := bytes.Repeat([]byte{1}, 59)

				channel := []byte{1, 2, 3, 4}

				initNoPad := append(channel, uint8(cmdMsg))
				initNoPad = append(initNoPad, []byte{0, 255}...)
				initNoPad = append(initNoPad, firstHalf...)
				init := zeroPad(initNoPad)

				contNoPad := append(channel, 0)
				contNoPad = append(contNoPad, secondHalf...)
				cont := zeroPad(contNoPad)

				// send init to allocate channel
				d, err := h.parseMsg(init)
				require.NoError(t, err)
				require.Nil(t, d)

				// first continuation packet, with sequence number 0
				d, err = h.parseMsg(cont)
				require.NoError(t, err)
				require.Nil(t, d)

				// send a continuation packet with sequence id 2, must error
				contNoPad = make([]byte, 64)
				contNoPad = append(channel, 2)
				contNoPad = append(contNoPad, secondHalf...)
				cont = zeroPad(contNoPad)

				d, err = h.handleContinuationPacket(cont)
				require.Error(t, err)
				require.Contains(t, err.Error(), "non-sequential sequence number")
				require.Nil(t, d)

				// send a continuation packet with sequence id 0 (already seen), must error
				contNoPad = make([]byte, 64)
				contNoPad = append(channel, 0)
				contNoPad = append(contNoPad, secondHalf...)
				cont = zeroPad(contNoPad)

				d, err = h.handleContinuationPacket(cont)
				require.Error(t, err)
				require.Contains(t, err.Error(), "non-sequential sequence number")
				require.Nil(t, d)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.f)
	}
}

func TestHandler_handleInitPacket(t *testing.T) {
	tests := []test{
		{
			"packet allocates completely and response can be built in one take",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				firstHalf := bytes.Repeat([]byte{1}, 57)

				channel := []byte{1, 2, 3, 4}

				initNoPad := append(channel, uint8(cmdMsg))
				initNoPad = append(initNoPad, []byte{0, 57}...)
				initNoPad = append(initNoPad, firstHalf...)
				init := zeroPad(initNoPad)

				// set bogus data for token to return
				token.shouldReturnData = true
				token.data = firstHalf
				d, err := h.handleInitPacket(init)
				require.NoError(t, err)
				require.NotNil(t, d)

				channelInt := binary.BigEndian.Uint32(channel)
				s, found := h.state.sessions[channelInt]
				require.NotNil(t, s)
				require.True(t, found)
				require.False(t, h.state.accumulatingMsgs)
			},
		},
		{
			"packet response cannot be built in one take",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				firstHalf := bytes.Repeat([]byte{1}, 57)

				channel := []byte{1, 2, 3, 4}

				initNoPad := append(channel, uint8(cmdMsg))
				initNoPad = append(initNoPad, []byte{0, 58}...)
				initNoPad = append(initNoPad, firstHalf...)
				init := zeroPad(initNoPad)

				d, err := h.handleInitPacket(init)
				require.NoError(t, err)
				require.Nil(t, d)

				channelInt := binary.BigEndian.Uint32(channel)
				s, found := h.state.sessions[channelInt]
				require.NotNil(t, s)
				require.True(t, found)
				require.True(t, h.state.accumulatingMsgs)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.f)
	}
}

func Test_numPackets(t *testing.T) {

	tests := []struct {
		name      string
		rawMsgLen int
		want      int
	}{
		{
			"59 bytes, single message",
			59,
			1,
		},
		{
			"0 bytes, no message",
			0,
			0,
		},
		{
			"2*59+1 bytes, 3 messages",
			119,
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, numPackets(tt.rawMsgLen))
		})
	}
}

func Test_broadcastReq(t *testing.T) {
	t.Run("initPacket command isn't cmdInit", func(t *testing.T) {
		i := initPacket{
			ChannelID: [4]byte{
				255,
				255,
				255,
				255,
			},
			Cmd:           cmdPing,
			PayloadLength: 0,
			Data:          nil,
		}

		d, err := broadcastReq(i)
		require.Nil(t, d)
		require.Error(t, err)
		require.Contains(t, err.Error(), "instead of U2FHID_INIT")
	})

	t.Run("initPacket command is cmdInit", func(t *testing.T) {
		i := initPacket{
			ChannelID: [4]byte{
				255,
				255,
				255,
				255,
			},
			Cmd:           cmdInit,
			PayloadLength: 8,
			Data:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		d, err := broadcastReq(i)
		require.NotNil(t, d)
		require.NoError(t, err)

		// channel is equal
		c := [4]byte{}
		copy(c[:], d[:4])
		require.Equal(t, i.ChannelID, c)

		// command is cmdInit
		require.Equal(t, i.Command(), d[4])

		// first byte of payload size is zero, second is 17
		require.Zero(t, d[5])
		require.Equal(t, uint8(17), d[6])

		// nonce is equal to what we put in data
		require.Equal(t, i.Data, d[7:7+8])

		// assigned channel id isn't all zeroes
		require.NotEqual(t, []byte{0, 0, 0, 0}, d[15:19])

		// remaining version bytes are equal
		require.Equal(t, uint8(12), d[19])
		require.Equal(t, uint8(4), d[20])
		require.Equal(t, uint8(2), d[21])
		require.Equal(t, uint8(0), d[22])
		require.Equal(t, uint8(0), d[23])
	})
}

func TestHandler_packetBuilder(t *testing.T) {
	tests := []test{
		{
			"unrecognized command",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				s := session{
					command: u2fHIDCommand(42),
				}

				dd, err := h.packetBuilder(&s, initPacket{})
				require.NotNil(t, dd)
				require.NoError(t, err)
				require.Len(t, dd, 1)

				d := dd[0]

				t.Log(d)

				require.Equal(t, []byte{0, 0, 0, 0}, d[:4])

				d = d[4:]
				require.Equal(t, uint8(cmdError), d[0])
				require.Equal(t, uint8(0), d[1])
				require.Equal(t, uint8(1), d[2])
			},
		},
		{
			"session says cmdInit, but pkt isn't initPacket",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				s := session{
					command: cmdInit,
				}

				dd, err := h.packetBuilder(&s, continuationPacket{})
				require.Nil(t, dd)
				require.Error(t, err)
				require.Contains(t, err.Error(), "said packet cannot be read as one")
			},
		},
		{
			"session says cmdInit, but channel isn't broadcast",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				s := session{
					command: cmdInit,
				}

				p := initPacket{
					ChannelID: [4]byte{1, 2, 3, 4},
				}

				dd, err := h.packetBuilder(&s, p)
				require.Nil(t, dd)
				require.Error(t, err)
				require.Contains(t, err.Error(), "not on the broadcast channel")
			},
		},
		{
			"session is cmdInit",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				s := session{
					command: cmdInit,
				}

				i := initPacket{
					ChannelID: [4]byte{
						255,
						255,
						255,
						255,
					},
					Cmd:           cmdInit,
					PayloadLength: 8,
					Data:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				}

				dd, err := h.packetBuilder(&s, i)
				require.NotNil(t, dd)
				require.NoError(t, err)
			},
		},
		{
			"session is cmdPing",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				i := initPacket{
					ChannelID: [4]byte{
						1,
						2,
						3,
						4,
					},
					Cmd:           cmdPing,
					PayloadLength: 8,
					Data:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				}

				s := session{
					command: cmdPing,
					data:    i.Data,
					total:   uint64(len(i.Data)),
				}

				dd, err := h.packetBuilder(&s, i)
				require.NoError(t, err)
				require.NotNil(t, dd)

				require.Len(t, dd, 1)

				d := dd[0][7:]

				d = bytes.Trim(d, "\x00")

				require.Equal(t, i.Data, d)
			},
		},
		{
			"session is cmdPing but something went wrong",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				i := initPacket{
					ChannelID: [4]byte{
						1,
						2,
						3,
						4,
					},
					Cmd:           cmdPing,
					PayloadLength: 8,
					Data:          []byte{1, 2, 3, 4, 5, 6, 7, 8},
				}

				s := session{
					command: cmdPing,
					total:   uint64(len(i.Data)),
				}

				dd, err := h.packetBuilder(&s, i)
				require.Error(t, err)
				require.Nil(t, dd)
			},
		},
		{
			"session is cmdMsg",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				firstHalf := bytes.Repeat([]byte{1}, 57)

				i := initPacket{
					ChannelID: [4]byte{
						1,
						2,
						3,
						4,
					},
					Cmd:           cmdMsg,
					PayloadLength: 57,
					Data:          firstHalf,
				}

				s := session{
					data:           firstHalf,
					command:        cmdMsg,
					total:          57,
					leftToRead:     0,
					lastSequence:   0,
					packetZeroSeen: false,
				}

				// set bogus data for token to return
				token.shouldReturnData = true
				token.data = firstHalf

				d, err := h.packetBuilder(&s, i)
				require.NoError(t, err)
				require.NotNil(t, d)
			},
		},
		{
			"session is cmdMsg but something goes wrong",
			func(t *testing.T) {
				token := &fakeToken{}
				h, err := NewHandler(token)
				require.NoError(t, err)

				firstHalf := bytes.Repeat([]byte{1}, 57)

				i := initPacket{
					ChannelID: [4]byte{
						1,
						2,
						3,
						4,
					},
					Cmd:           cmdMsg,
					PayloadLength: 57,
					Data:          firstHalf,
				}

				s := session{
					data:           firstHalf,
					command:        cmdMsg,
					total:          57,
					leftToRead:     0,
					lastSequence:   0,
					packetZeroSeen: false,
				}

				// set bogus data for token to return
				token.shouldReturnData = false
				token.data = firstHalf

				d, err := h.packetBuilder(&s, i)
				require.Error(t, err)
				require.Nil(t, d)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.f)
	}
}
