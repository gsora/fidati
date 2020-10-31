package storage_test

import (
	"crypto/rand"
	"testing"

	"github.com/gsora/fidati/storage"
	"github.com/stretchr/testify/require"
)

func Test_keyStorage_NewKeyItem(t *testing.T) {
	randData := func() []byte {
		r := make([]byte, 32)
		rand.Read(r)
		return r
	}

	tests := []struct {
		name    string
		appID   []byte
		wantErr bool
	}{
		{
			"creates a new keyItem successfully",
			randData(),
			false,
		},
		{
			"appID is nil",
			nil,
			true,
		},
		{
			"appID len == 0",
			[]byte{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := storage.New()
			require.NotNil(t, ks)

			ki, err := ks.NewKeyItem(tt.appID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, ki)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, ki)

			kki, err := ks.Item(ki.ID)
			require.NoError(t, err)
			require.NotEmpty(t, kki)
		})
	}
}

func Test_KeyStorage_Bytes(t *testing.T) {
	r := make([]byte, 32)
	rand.Read(r)
	ks := storage.New()
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
	ks := storage.New()
	_, err := ks.NewKeyItem(r)
	require.NoError(t, err)
	var bb []byte

	require.NotPanics(t, func() {
		bb = ks.Bytes()
	})

	require.NotNil(t, bb)

	// before testing LoadStorage(), we must assign a not-nil pointer to Storage
	storage.Storage = &storage.KeyStorage{}
	require.NotPanics(t, func() {
		err = storage.LoadStorage(bb)
	})

	require.NoError(t, err)
	require.Equal(t, ks, storage.Storage)
}

func Test_New(t *testing.T) {
	ks := storage.New()
	require.NotNil(t, ks)
	require.Empty(t, ks.M)
}

func TestKeyStorage_ItemExist(t *testing.T) {
	ks := storage.New()
	require.NotNil(t, ks)

	var sid [32]byte
	copy(sid[:], "appID")
	ks.M[sid] = storage.KeyItem{
		ID: sid,
	}

	ki, err := ks.Item(sid)

	require.NoError(t, err)
	require.Equal(t, ki, storage.KeyItem{
		ID: sid,
	})
}

func TestKeyStorage_ItemDoesntExist(t *testing.T) {
	ks := storage.New()
	require.NotNil(t, ks)

	var sid [32]byte
	copy(sid[:], "appID")
	ks.M[sid] = storage.KeyItem{
		ID: sid,
	}

	sid[6] = 'a'
	ki, err := ks.Item(sid)

	require.Error(t, err)
	require.Equal(t, ki, storage.KeyItem{})
}

func TestKeyStorage_IncrementKeyItemExist(t *testing.T) {
	ks := storage.New()
	require.NotNil(t, ks)

	var sid [32]byte
	copy(sid[:], "appID")
	ks.M[sid] = storage.KeyItem{
		ID: sid,
	}

	i, err := ks.IncrementKeyItem(sid)
	require.NoError(t, err)
	require.Equal(t, uint32(1), i)

	ki, err := ks.Item(sid)
	require.NoError(t, err)

	require.Equal(t, uint64(1), ki.Counter)
}

func TestKeyStorage_IncrementKeyItemDoesntExist(t *testing.T) {
	ks := storage.New()
	require.NotNil(t, ks)

	var sid [32]byte
	copy(sid[:], "appID")
	ks.M[sid] = storage.KeyItem{
		ID: sid,
	}

	sid[6] = 'a'

	i, err := ks.IncrementKeyItem(sid)
	require.Error(t, err)
	require.Equal(t, uint32(0), i)

	sid[6] = 0
	ki, err := ks.Item(sid)
	require.NoError(t, err)

	require.Equal(t, uint64(0), ki.Counter)
}
