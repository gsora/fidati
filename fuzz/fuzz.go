package fuzz

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/gsora/fidati/keyring"
	"github.com/gsora/fidati/u2fhid"
	"github.com/gsora/fidati/u2ftoken"
)

type fuzzCounter struct {
	counter uint32
}

func (f *fuzzCounter) Increment(appID []byte, challenge []byte, keyHandle []byte) (uint32, error) {
	f.counter++

	return f.counter, nil
}

func (f *fuzzCounter) UserPresence() bool {
	return true
}

func Fuzz(data []byte) int {
	l := &fuzzCounter{}
	k := keyring.New([]byte("key"), l)

	key, cert := genCert()
	token, err := u2ftoken.New(k, cert, key)
	if err != nil {
		panic("can't init u2ftoken")
	}

	hid, err := u2fhid.NewHandler(token)
	if err != nil {
		panic("can't init u2fhid")
	}

	out, err := hid.Rx(data, nil)
	if err != nil {
		if out != nil {
			panic("out != nil with err not nil")
		}
		return 0
	}

	return 1
}

func genCert() ([]byte, []byte) {
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

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return privkeyPem, certPem
}
