package u2ftoken

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

var errNotPEM = errors.New("not a PEM-encoded block")

func parseCert(c []byte) ([]byte, error) {
	certPem, _ := pem.Decode(c)
	if certPem == nil {
		return nil, errNotPEM
	}

	return certPem.Bytes, nil
}

func parseKey(k []byte) (*ecdsa.PrivateKey, error) {
	pkPem, _ := pem.Decode(k)
	if pkPem == nil {
		return nil, errNotPEM
	}

	attestationPrivkey, err := x509.ParseECPrivateKey(pkPem.Bytes)
	if err != nil {
		return nil, err
	}

	return attestationPrivkey, nil

}
