package u2fhid

// initResponse represents the standard response to a cmdInit command.
type initResponse struct {
	standardResponse
	Nonce              [8]byte
	AssignedChannelID  [4]byte
	ProtocolVersion    uint8
	MajorDeviceVersion uint8
	MinorDeviceVersion uint8
	BuildDeviceVersion uint8
	Capabilities       uint8
}

// standardResponse is the response header used by all response command.
type standardResponse struct {
	ChannelID [4]byte
	Command   uint8
	Count     [2]byte
}
