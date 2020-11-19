package attestation

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// ErrNotPem represents a PEM decoding error.
var ErrNotPEM = errors.New("not a PEM-encoded block")

// ParsePEM returns data block contained into c.
// Returns an error if c is not a PEM-encoded block.
func ParsePEM(c []byte) ([]byte, error) {
	certPem, _ := pem.Decode(c)
	if certPem == nil {
		return nil, ErrNotPEM
	}

	return certPem.Bytes, nil
}

// ParseKey parses a PEM-encoded ECDSA private key.
// Returns an error if k is not a PEM-encoded block, or the embedded block doesn't
// contain a valid ECDSA private key.
func ParseKey(k []byte) (*ecdsa.PrivateKey, error) {
	pkPem, _ := pem.Decode(k)
	if pkPem == nil {
		return nil, ErrNotPEM
	}

	attestationPrivkey, err := x509.ParseECPrivateKey(pkPem.Bytes)
	if err != nil {
		return nil, err
	}

	return attestationPrivkey, nil

}
