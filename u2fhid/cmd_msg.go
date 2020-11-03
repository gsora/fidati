package u2fhid

import (
	"log"
)

// handleMsg handles cmdMsg commands.
func (h *Handler) handleMsg(session *session, pkt u2fPacket) ([][]byte, error) {
	req, err := h.token.ParseRequest(session.data)
	if err != nil {
		return nil, err
	}

	log.Printf("%+v\n", req)

	resp := h.token.HandleMessage(req)

	return genPackets(resp, session.command, pkt.ChannelBytes())
}
