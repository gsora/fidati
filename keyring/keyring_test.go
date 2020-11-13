package keyring_test

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"testing"

	"github.com/gsora/fidati/keyring"
	"github.com/stretchr/testify/require"
)

type testCounter struct {
	i                 uint32
	userPresent       bool
	incrementMustFail bool
}

func (t *testCounter) Increment(appID []byte, challenge []byte, keyHandle []byte) (uint32, error) {
	if t.incrementMustFail {
		return 0, errors.New("cannot increment")
	}

	t.i++
	return t.i, nil
}

func (t *testCounter) UserPresence() bool {
	return t.userPresent
}

var nonceErr = func() ([]byte, error) {
	return nil, errors.New("nonce generation error")
}

var keygenErr = func(_ []byte) (*ecdsa.PrivateKey, error) {
	return nil, errors.New("key generation error")
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		mk      []byte
		counter keyring.Counter
	}{
		{
			"proper arguments (none nil)",
			[]byte("key"),
			&testCounter{},
		},
		{
			"nil key",
			nil,
			&testCounter{},
		},
		{
			"nil counter",
			[]byte("key"),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, keyring.New(tt.mk, tt.counter))
		})
	}
}

func TestKeyring_Register(t *testing.T) {
	k := keyring.New([]byte("key"), &testCounter{})

	tests := []struct {
		name       string
		kr         *keyring.Keyring
		appID      []byte
		wantErr    bool
		nonceFunc  keyring.NonceFuncType
		keygenFunc keyring.KeygenFuncType
	}{
		{
			"appID is not nil",
			k,
			[]byte("appID"),
			false,
			nil,
			nil,
		},
		{
			"appID is nil",
			k,
			nil,
			true,
			nil,
			nil,
		},
		{
			"nonce generation fails",
			k,
			[]byte("appID"),
			true,
			nonceErr,
			nil,
		},
		{
			"key generation fails",
			k,
			[]byte("appID"),
			true,
			nil,
			keygenErr,
		},
		{
			"keyring has nil master key",
			&keyring.Keyring{
				MasterKey: nil,
				Counter:   &testCounter{},
			},
			[]byte("appID"),
			true,
			nil,
			nil,
		},
		{
			"keyring has nil counter",
			&keyring.Keyring{
				MasterKey: []byte("key"),
				Counter:   nil,
			},
			[]byte("appID"),
			true,
			nil,
			nil,
		},
		{
			"keyring has both fields nil",
			&keyring.Keyring{
				MasterKey: nil,
				Counter:   nil,
			},
			[]byte("appID"),
			true,
			nil,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.nonceFunc != nil {
				oldNonceFunc := keyring.SetNonceFunc(tt.nonceFunc)
				defer func() {
					_ = keyring.SetNonceFunc(oldNonceFunc)
				}()
			}

			if tt.keygenFunc != nil {
				oldKeygenFunc := keyring.SetKeygenFunc(tt.keygenFunc)
				defer func() {
					_ = keyring.SetKeygenFunc(oldKeygenFunc)
				}()
			}

			pubKey, keyHandle, err := tt.kr.Register(tt.appID)

			if tt.wantErr {
				t.Log("error:", err)
				require.Error(t, err)
				require.Nil(t, pubKey)
				require.Nil(t, keyHandle)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, pubKey)
			require.NotNil(t, keyHandle)
		})
	}
}

func TestKeyring_RetrievePrivatekey(t *testing.T) {
	appID := bytes.Repeat([]byte{42}, 64)
	keyHandle := bytes.Repeat([]byte{43}, 64)
	masterKey := bytes.Repeat([]byte{44}, 64)

	firstIteration, err := keyring.RetrievePrivatekey(appID, keyHandle, masterKey)
	require.NoError(t, err)

	secondIteration, err := keyring.RetrievePrivatekey(appID, keyHandle, masterKey)
	require.NoError(t, err)

	require.True(t, firstIteration.Equal(secondIteration))
}

func TestKeyring_Authenticate(t *testing.T) {
	tc := &testCounter{}
	k := keyring.New([]byte("key"), tc)

	data := []byte("data")
	keyHandle := bytes.Repeat([]byte{42}, 64)

	type args struct {
		appID     []byte
		challenge []byte
		keyHandle []byte
	}

	tests := []struct {
		name           string
		args           args
		wantErr        bool
		incrementFails bool
		keygenFunc     keyring.KeygenFuncType
		kr             *keyring.Keyring
	}{
		{
			"appID is nil",
			args{
				nil,
				data,
				keyHandle,
			},
			true,
			false,
			nil,
			k,
		},
		{
			"challenge is nil",
			args{
				data,
				nil,
				keyHandle,
			},
			true,
			false,
			nil,
			k,
		},
		{
			"keyHandle is nil",
			args{
				data,
				data,
				nil,
			},
			true,
			false,
			nil,
			k,
		},
		{
			"keyHandle is less than 32 bytes long",
			args{
				data,
				data,
				data,
			},
			true,
			false,
			nil,
			k,
		},
		{
			"all fine but increment fails",
			args{
				data,
				data,
				keyHandle,
			},
			true,
			true,
			nil,
			k,
		},
		{
			"all fine but keygen function fails",
			args{
				data,
				data,
				keyHandle,
			},
			true,
			false,
			keygenErr,
			k,
		},
		{
			"all fine but keyring has nil master key",
			args{
				data,
				data,
				keyHandle,
			},
			true,
			false,
			nil,
			&keyring.Keyring{
				MasterKey: nil,
				Counter:   tc,
			},
		},
		{
			"all fine but keyring has nil counter",
			args{
				data,
				data,
				keyHandle,
			},
			true,
			false,
			nil,
			&keyring.Keyring{
				MasterKey: []byte("key"),
				Counter:   nil,
			},
		},
		{
			"all fine but keyring has nil counter and master key",
			args{
				data,
				data,
				keyHandle,
			},
			true,
			false,
			nil,
			&keyring.Keyring{
				MasterKey: nil,
				Counter:   nil,
			},
		},
		{
			"all fine",
			args{
				data,
				data,
				keyHandle,
			},
			false,
			false,
			nil,
			k,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc.incrementMustFail = tt.incrementFails

			if tt.keygenFunc != nil {
				oldKeygenFunc := keyring.SetKeygenFunc(tt.keygenFunc)
				defer func() {
					_ = keyring.SetKeygenFunc(oldKeygenFunc)
				}()
			}

			authSig, counter, err := tt.kr.Authenticate(tt.args.appID, tt.args.challenge, tt.args.keyHandle, true)

			if tt.wantErr {
				t.Log("error:", err)
				require.Error(t, err)
				require.Nil(t, authSig)
				require.Zero(t, counter)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, authSig)
			require.NotZero(t, counter)
		})
	}
}
