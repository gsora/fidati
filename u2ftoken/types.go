package u2ftoken

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"

	"github.com/gsora/fidati/attestation"
	"github.com/gsora/fidati/keyring"
)

// command represents a U2F standard command.
// See https://fidoalliance.org/specs/fido-u2f-v1.2-ps-20170411/fido-u2f-raw-message-formats-v1.2-ps-20170411.pdf for more
// details.
//go:generate stringer -type=command
type command uint8

const (
	_ command = iota
	// Register registers a new relying party.
	Register

	// Authenticate authenticates a relying party with the associated identity, stored in the token.
	Authenticate

	// Version returns the standard "U2F_V2" version string.
	Version
)

const (
	// The command completed successfully without error.
	noError errorCode = 0x9000

	// The request was rejected due to test-of-user-presence being required.
	errConditionNotSatisfied errorCode = 0x6985

	// The request was rejected due to an invalid key handle.
	errWrongData errorCode = 0x6A80

	// The length of the request was invalid.
	errWrongLength errorCode = 0x6700

	// The Class byte of the request is not supported.
	errClaNotSupported errorCode = 0x6E00

	// The Instruction of the request is not supported.
	errInsNotSupported errorCode = 0x6D00
)

// errorCode represents a U2F standard error code.
//go:generate stringer -type=errorCode
type errorCode uint16

// Error implements the error interface.
func (ec errorCode) Error() string {
	return ec.String()
}

// Bytes returns the byte array representation of c.
func (ec errorCode) Bytes() [2]byte {
	var ret [2]byte
	binary.BigEndian.PutUint16(ret[:], uint16(ec))
	return ret
}

// errorResponse returns a Response struct which wraps errCode.
func errorResponse(errCode errorCode) Response {
	return Response{
		Data:       []byte{},
		StatusCode: errCode.Bytes(),
	}
}

// Params holds the two APDU standard request parameters.
type Params struct {
	First  uint8
	Second uint8
}

// Request represents a standard APDU request.
type Request struct {
	Command          command
	Parameters       Params
	MaxResponseBytes uint16
	Data             []byte
}

// Response represents a standard APDU response.
type Response struct {
	Data       []byte
	StatusCode [2]byte
}

// Bytes returns the byte slice representation of r.
func (r Response) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.Data)
	buf.Write(r.StatusCode[:])
	return buf.Bytes()
}

// Token represents a U2F token.
// It handles request parsing and composition, key storage orchestration.
type Token struct {
	keyring                *keyring.Keyring
	attestationCertificate []byte
	attestationPrivkey     *ecdsa.PrivateKey
}

// New returns a new Token instance with k as Keyring.
// attCert must be a PEM-encoded certificate, while attPrivKey must be a X.509-encoded
// ECDSA private key.
func New(k *keyring.Keyring, attCert, attPrivKey []byte) (*Token, error) {
	cert, _, err := attestation.ParseCertificate(attCert)
	if err != nil {
		return nil, fmt.Errorf("cannot parse attestation certificate, %w", err)
	}

	key, err := attestation.ParseKey(attPrivKey)
	if err != nil {
		return nil, fmt.Errorf("cannot parse attestation certificate, %w", err)
	}

	return &Token{
		keyring:                k,
		attestationCertificate: cert,
		attestationPrivkey:     key,
	}, nil
}

// ParseRequest parses req as a U2F request.
// It returns a Request instance filled with the appropriate data from req, and an error.
func (t *Token) ParseRequest(req []byte) (Request, error) {
	var ret Request

	if req == nil {
		return Request{}, fmt.Errorf("request bytes are nil")
	}

	if req[0] != 0 {
		return Request{}, fmt.Errorf("first byte of request must be zero")
	}

	ret.Command = command(req[1])
	ret.Parameters = Params{
		First:  req[2],
		Second: req[3],
	}

	if req[4] != 0 {
		return Request{}, fmt.Errorf("fifth byte is not zero, must always be")
	}

	dataLen := binary.BigEndian.Uint16(req[5:7])

	if dataLen != 0 {
		ret.Data = req[7 : dataLen+7] // first 6 bytes are header tags, minus one for array indexing reasons :-)
	}

	// Ne initial offset = 6 (header bytes) + dataLen
	// Ne end offset = len(req)

	neBytes := req[(5 + dataLen):]

	if len(neBytes) == 3 {
		return Request{}, fmt.Errorf("Ne bytes are %d long while we were expecting 3 bytes", len(neBytes))
	}

	ret.MaxResponseBytes = binary.BigEndian.Uint16(neBytes)

	return ret, nil
}

// buildResponse returns a byte slice containing APDU bytes to appropriately respond to the associated Request.
func buildResponse(_ Request, resp Response) ([]byte, error) {
	return resp.Bytes(), nil

}
