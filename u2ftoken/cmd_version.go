package u2ftoken

import "fmt"

const versionPayload = "U2F_V2"

var versionString = []byte(fmt.Sprintf("%s", versionPayload))

func handleVersion(req Request) (Response, error) {
	return Response{
		Data:       versionString,
		StatusCode: NoError.Bytes(),
	}, nil
}
