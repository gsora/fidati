//go:generate go run github.com/rakyll/statik -src=../certs -p=certs

package u2ftoken

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"

	"github.com/gsora/fidati/storage"

	// statik generated files
	_ "github.com/gsora/fidati/u2ftoken/certs"
	"github.com/rakyll/statik/fs"
)

func init() {
	// initialize storage, empty one for now
	storage.Storage = storage.New()
	readCertPrivkey()
}

func readCertPrivkey() {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	aCert, err := statikFS.Open("/attestation_certificate.pem")
	if err != nil {
		panic(err)
	}

	aPk, err := statikFS.Open("/ecdsa_privkey.pem")
	if err != nil {
		panic(err)
	}

	aCertBytes, err := ioutil.ReadAll(aCert)
	if err != nil {
		panic(err)
	}

	aPkBytes, err := ioutil.ReadAll(aPk)
	if err != nil {
		panic(err)
	}

	certPem, _ := pem.Decode(aCertBytes)
	attestationCertificate = certPem.Bytes

	pkPem, _ := pem.Decode(aPkBytes)

	attestationPrivkey, err = x509.ParseECPrivateKey(pkPem.Bytes)
	if err != nil {
		panic(err)
	}
}
