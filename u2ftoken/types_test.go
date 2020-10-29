package u2ftoken_test

import (
	"testing"

	"github.com/gsora/fidati/u2ftoken"
	"github.com/stretchr/testify/require"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     []byte
		want    u2ftoken.Request
		wantErr bool
	}{
		{
			"a well-defined request",
			[]byte{0, 1, 3, 0, 0, 0, 64, 6, 9, 185, 109, 109, 182, 111, 97, 172, 47, 215, 18, 196, 234, 79, 162, 54, 190, 3, 167, 234, 133, 146, 237, 156, 188, 74, 123, 7, 141, 99, 125, 4, 100, 197, 158, 99, 237, 40, 17, 86, 119, 116, 14, 47, 119, 190, 29, 124, 182, 247, 109, 7, 163, 10, 193, 124, 139, 177, 78, 187, 207, 180, 205},
			u2ftoken.Request{
				Command: u2ftoken.Register,
				Parameters: u2ftoken.Params{
					First:  3,
					Second: 0,
				},
				Data: []byte{6, 9, 185, 109, 109, 182, 111, 97, 172, 47, 215, 18, 196, 234, 79, 162, 54, 190, 3, 167, 234, 133, 146, 237, 156, 188, 74, 123, 7, 141, 99, 125, 4, 100, 197, 158, 99, 237, 40, 17, 86, 119, 116, 14, 47, 119, 190, 29, 124, 182, 247, 109, 7, 163, 10, 193, 124, 139, 177, 78, 187, 207, 180, 205},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := u2ftoken.ParseRequest(tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, r)
		})
	}
}
