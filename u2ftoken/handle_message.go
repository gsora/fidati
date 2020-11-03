package u2ftoken

import (
	"errors"
	"log"
)

// a ready-made instance of the errConditionNotSatisfied error.
var notSatisfied = errorResponse(errConditionNotSatisfied).Bytes()

// HandleMessage handles a request, and returns a response byte slice.
func (t *Token) HandleMessage(req Request) []byte {
	var resp Response
	var handleErr error

	log.Printf("message type: %s\n", req.Command)

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
			// this is a strange error, log it and return ErrConditionNotSatisfied
			log.Println("non-u2f error detected:", handleErr)
			return notSatisfied
		}

		return errorResponse(err).Bytes()
	}

	log.Println("response len: ", len(resp.Bytes()))

	respBytes, err := buildResponse(req, resp)
	if err != nil {
		log.Println("cannot build response:", err)
		return notSatisfied
	}

	return respBytes
}
