package u2fhid

import (
	"log"

	"github.com/gsora/fidati/u2ftoken"
)

// handleMsg handles cmdMsg commands.
func handleMsg(session *session, pkt u2fPacket) ([][]byte, error) {
	req, err := u2ftoken.ParseRequest(session.data)
	if err != nil {
		return nil, err
	}

	log.Printf("%+v\n", req)

	resp := u2ftoken.HandleMessage(req)

	return genPackets(resp, session.command, pkt.ChannelBytes())
}
