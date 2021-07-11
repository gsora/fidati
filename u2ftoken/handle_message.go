package u2ftoken

import (
	"errors"

	"github.com/gsora/fidati/internal/flog"
)

// a ready-made instance of the errConditionNotSatisfied error.
var notSatisfied = errorResponse(errConditionNotSatisfied).Bytes()

// HandleMessage handles a message, and returns a response byte slice.
func (t *Token) HandleMessage(data []byte, _ uint32) []byte {
	req, err := t.ParseRequest(data)
	if err != nil {
		flog.Logger.Printf("cannot parse request, %s", err)
		return notSatisfied
	}

	var resp Response
	var handleErr error

	flog.Logger.Printf("request: %+v", req)

	switch req.Command {
	case Version:
		resp, handleErr = t.handleVersion(req)
	case Register:
		resp, handleErr = t.handleRegister(req)
	case Authenticate:
		resp, handleErr = t.handleAuthenticate(req)
	default:
		return notSatisfied
	}

	if handleErr != nil {
		var err errorCode

		if !errors.As(handleErr, &err) {
			// this is a strange error, flog.Logger.it and return ErrConditionNotSatisfied
			flog.Logger.Println("non-u2f error detected:", handleErr)
			return notSatisfied
		}

		return errorResponse(err).Bytes()
	}

	flog.Logger.Println("response len: ", len(resp.Bytes()))

	respBytes, err := buildResponse(req, resp)
	if err != nil {
		flog.Logger.Println("cannot build response:", err)
		return notSatisfied
	}

	return respBytes
}
