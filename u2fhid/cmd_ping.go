package u2fhid

// handlePing handles cmdPing commands.
func handlePing(session *session, pkt u2fPacket) ([][]byte, error) {
	// U2FHID_PING echoes back whatever you throw at it.
	return genPackets(session.data, session.command, pkt.ChannelBytes())
}
