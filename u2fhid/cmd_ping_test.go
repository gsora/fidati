package u2fhid

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_handlePing(t *testing.T) {
	t.Run("ping returns whatever you throw at it", func(t *testing.T) {
		s := &session{
			data:         bytes.Repeat([]byte{42}, 42),
			command:      cmdPing,
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
			Cmd:           cmdPing,
			PayloadLength: 42,
			Data:          bytes.Repeat([]byte{42}, 42),
		}

		data, err := handlePing(s, p)
		require.NoError(t, err)

		require.Len(t, data, 1)

		elem := data[0]
		require.Len(t, elem, 64)
		require.NotEmpty(t, elem)

		// last 57 bytes are equal to p.Data
		eqData := make([]byte, 57)
		copy(eqData, p.Data)
		require.Equal(t, eqData, elem[7:])
	})
}
