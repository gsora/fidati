package u2ftoken

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
)

const (
	// we expect no more than 64 bytes in the Data section of our Request.
	expectedDataLen = 64
)

func (t *Token) handleRegister(req Request) (Response, error) {
	if len(req.Data) != expectedDataLen {
		log.Printf("message length is %d instead of %d\n", len(req.Data), expectedDataLen)
		return Response{}, errWrongLength
	}

	challengeParam := req.Data[:32]
	appParam := req.Data[32:]

	newKey, err := t.storage.NewKeyItem(appParam)
	if err != nil {
		return Response{}, err
	}

	pubkey := elliptic.Marshal(elliptic.P256(), newKey.PrivateKey.PublicKey.X, newKey.PrivateKey.PublicKey.Y)

	resp := new(bytes.Buffer)

	resp.WriteByte(0x05)
	resp.Write(pubkey)

	resp.WriteByte(byte(len(newKey.ID)))

	resp.Write(newKey.ID[:])
	resp.Write(attestationCertificate)

	sigPayload := buildSigPayload(
		appParam,
		challengeParam,
		newKey.ID[:],
		pubkey,
	)

	sph := sha256.Sum256(sigPayload)
	spHash := sph[:]

	sign, err := ecdsa.SignASN1(rand.Reader, attestationPrivkey, spHash)
	if err != nil {
		return Response{}, err
	}

	log.Println("sign len:", len(sign))
	resp.Write(sign)
	return Response{
		Data:       resp.Bytes(),
		StatusCode: noError.Bytes(),
	}, nil
}

func buildSigPayload(appParam []byte, challenge []byte, key []byte, pubKey []byte) []byte {
	p := new(bytes.Buffer)
	p.WriteByte(0x00)
	p.Write(appParam)
	p.Write(challenge)
	p.Write(key)
	p.Write(pubKey)

	return p.Bytes()
}
