package u2ftoken

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/gsora/fidati/internal/flog"
)

const (
	// we expect no less than minimumLen bytes when parsing an Authenticate request.
	minimumLen = 32 + 32 + 1 // control byte + challenge param + app param + key handle len

	controlCheckOnly                      = 0x07
	controlEnforceUserPresenceAndSign     = 0x03
	controlDontEnforceUserPresenceAndSign = 0x08
)

func (t *Token) handleAuthenticate(req Request) (Response, error) {
	if len(req.Data) < minimumLen {
		return Response{}, errWrongLength
	}

	controlByte := req.Parameters.First

	flog.Logger.Println("control byte ", controlByte)

	flog.Logger.Println("data", hex.EncodeToString(req.Data))
	challengeParam := req.Data[0:32]
	appID := req.Data[32:64]
	khLen := req.Data[64]

	flog.Logger.Printf("challenge len %d, app param len %d, khlen %d, total len data %d\n", len(challengeParam), len(appID), khLen, len(req.Data))

	if len(req.Data) != int(minimumLen+khLen) {
		flog.Logger.Printf("len request data %d different from minimumLen+khLen %d", len(req.Data), int(minimumLen+khLen))
		// total data len must be equal to minimumLen + khLen (headers + length of the key handle)
		return Response{}, errWrongLength
	}

	keyHandle := req.Data[minimumLen : minimumLen+khLen]

	flog.Logger.Println("requesting appID:", hex.EncodeToString(appID))
	flog.Logger.Println("requesting keyHandle:", hex.EncodeToString(keyHandle))

	// check that appID derives the same keyHandle we received
	nonce := t.keyring.NonceFromKeyHandle(keyHandle)
	if nonce == nil {
		flog.Logger.Println("cannot obtain nonce from provided key handle")
		return Response{}, errWrongData
	}

	_, derivedKeyHandle, err := t.keyring.Register(appID, nonce)
	if err != nil {
		flog.Logger.Println("cannot register key:", err)
		return Response{}, errWrongData
	}

	flog.Logger.Println("generated keyhandle from appID:", hex.EncodeToString(derivedKeyHandle))

	if !bytes.Equal(derivedKeyHandle, keyHandle) {
		flog.Logger.Println("derived key handle and provided key handle don't match")
		return Response{}, errWrongData
	}

	userPresence := t.keyring.Counter.UserPresence()

	// we only handle those two cases because the last one basically means
	// "authenticate, thanks"
	switch controlByte {
	case controlCheckOnly:
		return Response{}, errConditionNotSatisfied
	case controlEnforceUserPresenceAndSign:
		if !userPresence {
			flog.Logger.Println("control byte asked to enforce user presence, but it wasn't present")
			return Response{}, errConditionNotSatisfied
		}
	}

	userPresenceByte := byte(0)
	if userPresence {
		userPresenceByte = 1
	}

	sign, ni, err := t.keyring.Authenticate(appID, challengeParam, keyHandle, userPresence)
	if err != nil {
		return Response{}, errWrongData
	}

	resp := new(bytes.Buffer)
	resp.WriteByte(userPresenceByte)

	counterBytes := [4]byte{}
	binary.BigEndian.PutUint32(counterBytes[:], ni)

	resp.Write(counterBytes[:])
	resp.Write(sign)

	return Response{
		Data:       resp.Bytes(),
		StatusCode: noError.Bytes(),
	}, nil
}
