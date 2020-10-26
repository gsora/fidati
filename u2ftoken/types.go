package u2ftoken

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

var (
	ulog = log.New(os.Stdout, "u2ftoken :: ", log.Lshortfile)

	attestationCertificate []byte
	attestationPrivkey     *ecdsa.PrivateKey
)

//go:generate stringer -type=Command
type Command uint8

const (
	_ Command = iota
	Register
	Authenticate
	Version
)

/*
SW_NO_ERROR (0x9000): The command completed successfully without error.
SW_CONDITIONS_NOT_SATISFIED (0x6985): The request was rejected due to test-of-user-presence being required.
SW_WRONG_DATA (0x6A80): The request was rejected due to an invalid key handle.
SW_WRONG_LENGTH (0x6700): The length of the request was invalid.
SW_CLA_NOT_SUPPORTED (0x6E00): The Class byte of the request is not supported.
SW_INS_NOT_SUPPORTED (0x6D00): The Instruction of the request is not supported
*/

//go:generate stringer -type=ErrorCode
type ErrorCode uint16

func (ec ErrorCode) Error() string {
	return ec.String()
}

func (c ErrorCode) Bytes() [2]byte {
	var ret [2]byte
	binary.BigEndian.PutUint16(ret[:], uint16(c))
	return ret
}

const (
	NoError                  ErrorCode = 0x9000
	ErrConditionNotSatisfied ErrorCode = 0x6985
	ErrWrongData             ErrorCode = 0x6A80
	ErrWrongLength           ErrorCode = 0x6700
	ErrClaNotSupported       ErrorCode = 0x6E00
	ErrInsNotSupported       ErrorCode = 0x6D00
)

func ErrorResponse(errCode ErrorCode) Response {
	return Response{
		Data:       []byte{},
		StatusCode: errCode.Bytes(),
	}
}

type Params struct {
	First  uint8
	Second uint8
}

type Request struct {
	Command          Command
	Parameters       Params
	MaxResponseBytes uint16
	Data             []byte
}

type Response struct {
	Data       []byte
	StatusCode [2]byte
}

func (r Response) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, r.Data)
	buf.Write(r.StatusCode[:])
	return buf.Bytes()
}

func ParseRequest(req []byte) (Request, error) {
	var ret Request

	if req == nil {
		return Request{}, fmt.Errorf("request bytes are nil")
	}

	if req[0] != 0 {
		return Request{}, fmt.Errorf("first byte of request must be zero")
	}

	ret.Command = Command(req[1])
	ret.Parameters = Params{
		First:  req[2],
		Second: req[3],
	}

	if req[4] != 0 {
		return Request{}, fmt.Errorf("fifth byte is not zero, must always be")
	}

	dataLen := binary.BigEndian.Uint16(req[5:7])

	if dataLen != 0 {
		ret.Data = req[7 : dataLen+7] // first 6 bytes are header tags, minus one for array indexing reasons :-)
	}

	// Ne initial offset = 6 (header bytes) + dataLen
	// Ne end offset = len(req)

	neBytes := req[(5 + dataLen):]

	if len(neBytes) == 3 {
		panic(fmt.Sprintf("Ne bytes are %d long while we were expecting 3 bytes", len(neBytes)))
	}

	ret.MaxResponseBytes = binary.BigEndian.Uint16(neBytes)

	return ret, nil
}

func buildResponse(req Request, resp Response) ([]byte, error) {
	/*if len(resp.Data) > int(req.MaxResponseBytes) && req.Command != Version {
		return ErrorResponse(ErrWrongLength).Bytes(), fmt.Errorf("data is longer than max response bytes requested")
	}*/

	return resp.Bytes(), nil

}
