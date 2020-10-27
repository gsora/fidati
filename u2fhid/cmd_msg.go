package u2fhid

import "github.com/f-secure-foundry/tamago-example/u2ftoken"

// handleMsg handles cmdMsg commands.
func handleMsg(session *session, pkt u2fPacket) ([][]byte, error) {
	req, err := u2ftoken.ParseRequest(session.data)
	if err != nil {
		return nil, err
	}

	ulog.Printf("%+v\n", req)

	resp := u2ftoken.HandleMessage(req)

	return genPackets(resp, session.command, pkt.ChannelBytes())
}
