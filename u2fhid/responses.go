package u2fhid

type initResponse struct {
	ChannelID          [4]byte
	Command            uint8
	Count              [2]byte
	Nonce              [8]byte
	ProtocolVersion    uint8
	MajorDeviceVersion uint8
	MinorDeviceVersion uint8
	BuildDeviceVersion uint8
	Capabilities       uint8
}

// len = 7
type pingResponse struct {
	ChannelID [4]byte
	Command   uint8
	Count     [2]byte
}

// len = 7
type msgResponse struct {
	ChannelID [4]byte
	Command   uint8
	Count     [2]byte
}

type standardResponse struct {
	ChannelID [4]byte
	Command   uint8
	Count     [2]byte
}
