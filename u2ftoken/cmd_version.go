package u2ftoken

import "fmt"

// this is the standard response a U2F relying party expects when it sends a
// Version command.
const versionPayload = "U2F_V2"

var versionString = []byte(fmt.Sprintf("%s", versionPayload))

func handleVersion(req Request) (Response, error) {
	return Response{
		Data:       versionString,
		StatusCode: noError.Bytes(),
	}, nil
}
