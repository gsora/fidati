package u2ftoken

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_keyStorage_newKeyItem(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"creates a new keyItem successfully",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := make([]byte, 32)
			rand.Read(r)
			ks := newKeyStorage()
			ki, err := ks.newKeyItem(r)
			require.NoError(t, err)

			kki, err := ks.item(ki.ID)
			require.NoError(t, err)
			require.NotEmpty(t, kki)
		})
	}
}
