package u2ftoken

import (
	"bytes"
	"encoding/binary"

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

	challengeParam := req.Data[0:32]
	appID := req.Data[32:64]
	khLen := req.Data[64]

	flog.Logger.Printf("challenge len %d, app param len %d, khlen %d", len(challengeParam), len(appID), khLen)

	if len(req.Data) != int(minimumLen+khLen) {
		flog.Logger.Printf("len request data %d different from minimumLen+khLen %d", len(req.Data), int(minimumLen+khLen))
		// total data len must be equal to minimumLen + khLen (headers + length of the key handle)
		return Response{}, errWrongLength
	}

	keyHandle := req.Data[minimumLen : minimumLen+khLen]

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
