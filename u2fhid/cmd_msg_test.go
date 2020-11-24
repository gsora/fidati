package u2fhid

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_handleMsg(t *testing.T) {

	tests := []struct {
		name             string
		token            Token
		errAssertion     require.ErrorAssertionFunc
		packetsAssertion require.ValueAssertionFunc
	}{
		{
			"underlying token returns no error",
			&fakeToken{
				shouldReturnData: true,
				data:             []byte("data"),
			},
			require.NoError,
			require.NotEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewHandler(tt.token)
			require.NoError(t, err)

			s := &session{
				data:         bytes.Repeat([]byte{42}, 42),
				command:      cmdMsg,
				total:        42,
				leftToRead:   0,
				lastSequence: 0,
			}

			p := initPacket{
				ChannelID: [4]byte{
					1,
					2,
					3,
					4,
				},
				Cmd:           cmdMsg,
				PayloadLength: 42,
				Data:          bytes.Repeat([]byte{42}, 42),
			}

			data, err := u.handleMsg(s, p)
			tt.errAssertion(t, err)
			tt.packetsAssertion(t, data)
		})
	}
}
