package u2fhid

import "github.com/gsora/fidati/internal/flog"

// handleMsg handles cmdMsg commands.
func (h *Handler) handleMsg(session *session, pkt u2fPacket) ([][]byte, error) {
	req, err := h.token.ParseRequest(session.data)
	if err != nil {
		return nil, err
	}

	flog.Logger.Printf("%+v\n", req)

	resp := h.token.HandleMessage(req)

	return genPackets(resp, session.command, pkt.ChannelBytes())
}
