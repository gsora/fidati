package storage_test

import (
	"crypto/rand"
	"errors"
	"io"
	"testing"

	"github.com/gsora/fidati/storage"
	"github.com/stretchr/testify/require"
)

// failRW is a io.ReadWriter which always fails.
type failRW struct{}

func (f *failRW) Read(p []byte) (n int, err error) {
	return -1, errors.New("error")
}

func (f *failRW) Write(p []byte) (n int, err error) {
	return -1, errors.New("error")
}

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
		device  io.ReadWriter
	}{
		{
			"creates a new keyItem successfully",
			randData(),
			false,
			nil,
		},
		{
			"appID is nil",
			nil,
			true,
			nil,
		},
		{
			"appID len == 0",
			[]byte{},
			true,
			nil,
		},
		{
			"creates a new keyItem successfully but Device write returns error",
			randData(),
			true,
			&failRW{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.device != nil {
				storage.Device = tt.device
				defer func() {
					storage.Device = storage.NilDevice{}
				}()
			}
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

func TestKeyStorage_IncrementKeyItemExistDeviceFails(t *testing.T) {
	storage.Device = &failRW{}

	ks := storage.New()
	require.NotNil(t, ks)

	var sid [32]byte
	copy(sid[:], "appID")
	ks.M[sid] = storage.KeyItem{
		ID: sid,
	}

	i, err := ks.IncrementKeyItem(sid)
	require.Error(t, err)
	require.Equal(t, uint32(0), i)

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
