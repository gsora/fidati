package u2ftoken

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"log"
)

const (
	// we expect no less than minimumLen bytes when parsing an Authenticate request.
	minimumLen = 32 + 32 + 1 // control byte + challenge param + app param + key handle len

	controlCheckOnly                      = 0x07
	controlEnforceUserPresenceAndSign     = 0x03
	controlDontEnforceUserPresenceAndSign = 0x08
)

func handleAuthenticate(req Request) (Response, error) {
	if len(req.Data) < minimumLen {
		return Response{}, errWrongLength
	}

	controlByte := req.Parameters.First

	log.Println("control byte ", controlByte)

	log.Printf("keys %+v", ks.M)

	challengeParam := req.Data[0:32]
	appParam := req.Data[32:64]
	khLen := req.Data[64]

	log.Printf("challenge len %d, app param len %d, khlen %d", len(challengeParam), len(appParam), khLen)

	if len(req.Data) != int(minimumLen+khLen) {
		log.Printf("len request data %d different from minimumLen+khLen %d", len(req.Data), int(minimumLen+khLen))
		// total data len must be equal to minimumLen + khLen (headers + length of the key handle)
		return Response{}, errWrongLength
	}

	kh := req.Data[minimumLen : minimumLen+khLen]
	var khKey [32]byte
	copy(khKey[:], kh)

	ki, err := ks.item(khKey)
	log.Println("query item ", ki, err)

	if controlByte == controlCheckOnly {
		if err == nil { // key found
			log.Println("key found")
			return Response{
				StatusCode: errConditionNotSatisfied.Bytes(),
			}, nil
		}

		return Response{
			StatusCode: errWrongData.Bytes(),
		}, nil
	} else if err != nil {
		log.Println("some kind of error", err)
		return Response{}, err
	}

	ni, err := ks.incrementKeyItem(khKey)
	if err != nil {
		return Response{}, err
	}

	sp := signaturePayload(
		appParam,
		ni,
		challengeParam,
	)

	sph := sha256.Sum256(sp)
	spHash := sph[:]

	sign, err := ecdsa.SignASN1(rand.Reader, ki.PrivateKey, spHash)
	if err != nil {
		return Response{}, err
	}

	resp := new(bytes.Buffer)
	resp.WriteByte(1) // user presence

	counterBytes := [4]byte{}
	binary.BigEndian.PutUint32(counterBytes[:], ni)

	resp.Write(counterBytes[:])
	resp.Write(sign)

	return Response{
		Data:       resp.Bytes(),
		StatusCode: noError.Bytes(),
	}, nil
}

func signaturePayload(appParam []byte, counter uint32, challengeParam []byte) []byte {
	ret := new(bytes.Buffer)

	ret.Write(appParam)
	ret.WriteByte(1) // assume we checked user presence for now

	counterBytes := [4]byte{}

	binary.BigEndian.PutUint32(counterBytes[:], counter)

	ret.Write(counterBytes[:])
	ret.Write(challengeParam)

	return ret.Bytes()
}
