package storage

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
			ks := New()
			ki, err := ks.NewKeyItem(r)
			require.NoError(t, err)

			kki, err := ks.Item(ki.ID)
			require.NoError(t, err)
			require.NotEmpty(t, kki)
		})
	}
}

func Test_keyStorage_Bytes(t *testing.T) {
	r := make([]byte, 32)
	rand.Read(r)
	ks := New()
	_, err := ks.NewKeyItem(r)
	require.NoError(t, err)
	var bb []byte

	require.NotPanics(t, func() {
		bb = ks.Bytes()
	})

	require.NotNil(t, bb)
}

func Test_LoadStorage(t *testing.T) {
	r := make([]byte, 32)
	rand.Read(r)
	ks := New()
	_, err := ks.NewKeyItem(r)
	require.NoError(t, err)
	var bb []byte

	require.NotPanics(t, func() {
		bb = ks.Bytes()
	})

	require.NotNil(t, bb)

	// before testing LoadStorage(), we must assign a not-nil pointer to Storage
	Storage = &KeyStorage{}
	require.NotPanics(t, func() {
		err = LoadStorage(bb)
	})

	require.NoError(t, err)
	require.Equal(t, ks, Storage)
}
