package u2fhid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_u2fHIDReport_Bytes(t *testing.T) {
	t.Run("hid report bytes are serialized", func(t *testing.T) {
		require.NotNil(t, DefaultReport.Bytes())
	})
}

func TestNewHandler(t *testing.T) {

	tests := []struct {
		name          string
		token         Token
		errAssertion  require.ErrorAssertionFunc
		dataAssertion require.ValueAssertionFunc
	}{
		{
			"token is not nil",
			&fakeToken{},
			require.NoError,
			require.NotNil,
		},
		{
			"token is nil",
			nil,
			require.Error,
			require.Nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHandler(tt.token)
			tt.errAssertion(t, err)
			tt.dataAssertion(t, got)
		})
	}
}

func Test_session_clear(t *testing.T) {
	t.Run("state values are set to their default values", func(t *testing.T) {
		s := &session{
			data:         []byte("some data"),
			command:      cmdMsg,
			total:        42,
			leftToRead:   0,
			lastSequence: 42,
		}

		s.clear()

		require.Equal(t, session{}, *s)
	})
}

func Test_u2fHIDState_clear(t *testing.T) {
	t.Run("state values are set to their default values", func(t *testing.T) {
		u := &u2fHIDState{
			outboundMsgs: [][]byte{
				[]byte("some data"),
			},
			lastOutboundIndex: 42,
			accumulatingMsgs:  true,
			sessions: map[uint32]*session{
				42: {
					data:         []byte("some data"),
					command:      cmdMsg,
					total:        42,
					leftToRead:   0,
					lastSequence: 42,
				},
			},
			lastChannelID: 42,
		}

		u.clear()

		require.Nil(t, u.outboundMsgs)
		require.Zero(t, u.lastOutboundIndex)
		require.False(t, u.accumulatingMsgs)
		require.NotNil(t, u.sessions[42])
		require.Len(t, u.sessions, 1)
		require.Zero(t, u.lastChannelID)
	})
}
