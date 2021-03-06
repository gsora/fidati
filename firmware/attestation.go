//go:generate go run github.com/rakyll/statik -src=../certs -p=certs

package main

import (
	"io/ioutil"

	"github.com/rakyll/statik/fs"
)

var (
	// X.509 attestation certificate, sent along in registration requests
	attestationCertificate []byte

	// ECDSA private key, used to sign registration requests
	attestationPrivkey []byte
)

func readCertPrivkey() {
	statikFS, err := fs.New()
	notErr(err)

	aCert, err := statikFS.Open("/attestation_certificate.pem")
	notErr(err)

	aPk, err := statikFS.Open("/ecdsa_privkey.pem")
	notErr(err)

	aCertBytes, err := ioutil.ReadAll(aCert)
	notErr(err)

	aPkBytes, err := ioutil.ReadAll(aPk)
	notErr(err)

	attestationCertificate = aCertBytes
	attestationPrivkey = aPkBytes
}
