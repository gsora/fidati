package u2ftoken

import "errors"

var notSatisfied = ErrorResponse(ErrConditionNotSatisfied).Bytes()

func HandleMessage(req Request) []byte {
	var resp Response
	var handleErr error

	ulog.Printf("message type: %s\n", req.Command)

	switch req.Command {
	case Version:
		resp, handleErr = handleVersion(req)
	case Register:
		resp, handleErr = handleRegister(req)
	case Authenticate:
		resp, handleErr = handleAuthenticate(req)
	default:
		return notSatisfied
	}

	if handleErr != nil {
		var err ErrorCode

		if !errors.As(handleErr, &err) {
			// this is a strange error, log it and return ErrConditionNotSatisfied
			ulog.Println("non-u2f error detected:", handleErr)
			return notSatisfied
		}

		return ErrorResponse(err).Bytes()
	}

	ulog.Println("response len: ", len(resp.Bytes()))

	respBytes, err := buildResponse(req, resp)
	if err != nil {
		ulog.Println("cannot build response:", err)
		return notSatisfied
	}

	return respBytes
}
