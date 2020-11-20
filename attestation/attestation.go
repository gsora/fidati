package attestation

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// ErrNotPem represents a PEM decoding error.
var ErrNotPEM = errors.New("not a PEM-encoded block")

// ParseCertificate parses a X.509 certificate from the data contained in the
// PEM data block.
// Returns an error when c is not valid PEM data, or a valid X.509 certificate.
func ParseCertificate(c []byte) ([]byte, *x509.Certificate, error) {
	certPem, _ := pem.Decode(c)
	if certPem == nil {
		return nil, nil, ErrNotPEM
	}

	cert, err := x509.ParseCertificate(certPem.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid X.509 certificate, %w", err)
	}

	return certPem.Bytes, cert, nil
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
