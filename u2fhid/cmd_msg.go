package u2fhid

// handleMsg handles cmdMsg commands.
func (h *Handler) handleMsg(session *session, pkt u2fPacket) ([][]byte, error) {
	return genPackets(
		h.token.HandleMessage(session.data[:session.total]),
		session.command,
		pkt.ChannelBytes(),
	)
}
