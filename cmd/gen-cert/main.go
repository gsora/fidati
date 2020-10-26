package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"time"
)

func main() {
	serial, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		log.Fatal("cannot generate serial, ", err)
	}
	cert := &x509.Certificate{
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0), // this certificate is valid for 10 years
		SerialNumber: serial,
		Issuer: pkix.Name{
			Country:      []string{"IT"},
			SerialNumber: "",
			CommonName:   "Fidati U2F Token",
		},
		PublicKeyAlgorithm: x509.ECDSA,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}

	cert.Subject = cert.Issuer

	privkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal("cannot generate ecdsa key ", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privkey.PublicKey, privkey)
	if err != nil {
		log.Fatal("cannot generate certificate, ", err)
	}

	privkeyDer, err := x509.MarshalECPrivateKey(privkey)
	if err != nil {
		log.Fatal("cannot marshal ecdsa privkey, ", err)
	}

	privkeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privkeyDer,
	})

	err = ioutil.WriteFile("ecdsa_privkey.pem", privkeyPem, 0644)
	if err != nil {
		log.Fatal("cannot write private key: ", err)
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	err = ioutil.WriteFile("attestation_certificate.pem", certPem, 0644)
	if err != nil {
		log.Fatal("cannot write certificate: ", err)
	}

}
