package u2ftoken

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"testing"

	"github.com/gsora/fidati/keyring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_errorCode_Error(t *testing.T) {
	t.Run("errorCode implements error interface", func(t *testing.T) {
		var err error
		err = errWrongData

		require.True(t, errors.Is(err, errWrongData))

		require.Equal(t, "errWrongData", err.Error())
	})
}

func Test_errorCode_Bytes(t *testing.T) {
	t.Run("a uint16 is represented by a slice of 2 uint8", func(t *testing.T) {
		expected := [2]byte{0x6D, 0x00}

		require.Equal(t, expected, errInsNotSupported.Bytes())
	})
}

func Test_errorResponse(t *testing.T) {
	t.Run("errorResponse returns empty data, and a status code", func(t *testing.T) {
		r := errorResponse(noError)
		expectedErrCode := [2]byte{0x90, 0x00}

		require.Empty(t, r.Data)
		require.Equal(t, expectedErrCode, r.StatusCode)
	})
}

func TestResponse_Bytes(t *testing.T) {
	t.Run("bytes returns the byte representation of Response, with data bytes before and error code after",
		func(t *testing.T) {
			r := Response{
				Data:       bytes.Repeat([]byte{1}, 42),
				StatusCode: noError.Bytes(),
			}

			var out []byte
			require.NotPanics(t, func() {
				out = r.Bytes()
			})

			require.Equal(t, r.Data, out[:42])
			require.Equal(t, r.StatusCode[:], out[42:])
		})
}

func TestNew(t *testing.T) {
	type args struct {
		k          *keyring.Keyring
		attCert    []byte
		attPrivKey []byte
	}
	tests := []struct {
		name      string
		args      args
		want      *Token
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.k, tt.args.attCert, tt.args.attPrivKey)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToken_ParseRequest(t *testing.T) {
	type fields struct {
		keyring                *keyring.Keyring
		attestationCertificate []byte
		attestationPrivkey     *ecdsa.PrivateKey
	}
	type args struct {
		req []byte
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      Request
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Token{
				keyring:                tt.fields.keyring,
				attestationCertificate: tt.fields.attestationCertificate,
				attestationPrivkey:     tt.fields.attestationPrivkey,
			}
			got, err := tr.ParseRequest(tt.args.req)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_buildResponse(t *testing.T) {
	type args struct {
		in0  Request
		resp Response
	}
	tests := []struct {
		name      string
		args      args
		want      []byte
		assertion assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildResponse(tt.args.in0, tt.args.resp)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
