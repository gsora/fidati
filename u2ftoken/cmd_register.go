package u2ftoken

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gsora/fidati/internal/flog"
)

const (
	// we expect no more than 64 bytes in the Data section of our Request.
	expectedDataLen = 64
)

func (t *Token) handleRegister(req Request) (Response, error) {
	if len(req.Data) != expectedDataLen {
		flog.Logger.Printf("message length is %d instead of %d\n", len(req.Data), expectedDataLen)
		return Response{}, errWrongLength
	}

	if !t.keyring.Counter.UserPresence() {
		flog.Logger.Println("user presence during registration is required")
		return Response{}, errConditionNotSatisfied
	}

	challengeParam := req.Data[:32]
	appID := req.Data[32:]

	newKey, keyHandle, err := t.keyring.Register(appID, nil)
	if err != nil {
		return Response{}, err
	}

	pubkey := elliptic.Marshal(elliptic.P256(), newKey.X, newKey.Y)

	resp := new(bytes.Buffer)

	resp.WriteByte(0x05)
	resp.Write(pubkey)

	resp.WriteByte(byte(len(keyHandle)))

	resp.Write(keyHandle)
	resp.Write(t.attestationCertificate)

	sigPayload := buildSigPayload(
		appID,
		challengeParam,
		keyHandle,
		pubkey,
	)

	flog.Logger.Println("data", hex.EncodeToString(req.Data))
	flog.Logger.Println("registered appID:", hex.EncodeToString(appID))
	flog.Logger.Println("registered keyhandle:", hex.EncodeToString(keyHandle))

	sph := sha256.Sum256(sigPayload)
	spHash := sph[:]

	sign, err := ecdsa.SignASN1(rand.Reader, t.attestationPrivkey, spHash)
	if err != nil {
		return Response{}, err
	}

	flog.Logger.Println("sign len:", len(sign))
	resp.Write(sign)

	rb := resp.Bytes()
	flog.Logger.Println("response bytes:", hex.EncodeToString(rb))

	return Response{
		Data:       rb,
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
