package u2fhid

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Token represents a unit which can handle U2F messages.
type Token interface {
	// HandleMessage handles cmdMsg payloads, and return an appropriate response
	// for the underlying command.
	HandleMessage([]byte) []byte
}

// u2fHIDReport is a byte slice holding a standard U2F HID report.
type u2fHIDReport []byte

// Bytes returns the byte slice representation of r.
func (r *u2fHIDReport) Bytes() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, r)
	if err != nil {
		panic(fmt.Sprintf("cannot format u2f hid report, %v", err))
	}
	return buf.Bytes()
}

// DefaultReport is the standard report descriptor for a USB HID FIDO2 token.
// https://chromium.googlesource.com/chromiumos/platform2/+/master/u2fd/u2fhid.cc
var DefaultReport = u2fHIDReport{
	0x06, 0xD0, 0xF1, /* Usage Page (FIDO Alliance), FIDO_USAGE_PAGE */
	0x09, 0x01, /* Usage (U2F HID Auth. Device) FIDO_USAGE_U2FHID */
	0xA1, 0x01, /* Collection (Application), HID_APPLICATION */
	0x09, 0x20, /*  Usage (Input Report Data), FIDO_USAGE_DATA_IN */
	0x15, 0x00, /*  Logical Minimum (0) */
	0x26, 0xFF, 0x00, /*  Logical Maximum (255) */
	0x75, 0x08, /*  Report Size (8) */
	0x95, 0x40, /*  Report Count (64), HID_INPUT_REPORT_BYTES */
	0x81, 0x02, /*  Input (Data, Var, Abs), Usage */
	0x09, 0x21, /*  Usage (Output Report Data), FIDO_USAGE_DATA_OUT */
	0x15, 0x00, /*  Logical Minimum (0) */
	0x26, 0xFF, 0x00, /*  Logical Maximum (255) */
	0x75, 0x08, /*  Report Size (8) */
	0x95, 0x40, /*  Report Count (64), HID_OUTPUT_REPORT_BYTES */
	0x91, 0x02, /*  Output (Data, Var, Abs), Usage */
	0xC0, /* End Collection */
}

//go:generate stringer -type=u2fHIDCommand
type u2fHIDCommand int

const (
	broadcastChan = 0xffffffff

	// mandatory commands
	cmdPing  u2fHIDCommand = 0x80 | 0x01
	cmdMsg   u2fHIDCommand = 0x80 | 0x03
	cmdInit  u2fHIDCommand = 0x80 | 0x06
	cmdError u2fHIDCommand = 0x80 | 0x3f

	// optional commands
	cmdLock u2fHIDCommand = 0x80 | 0x04
	cmdWink u2fHIDCommand = 0x80 | 0x08
	cmdSync u2fHIDCommand = 0x80 | 0x3c
)

// Handler holds methods for sending and receiving packets.
type Handler struct {
	// token instance
	token Token

	// u2fhid state
	state *u2fHIDState
}

// NewHandler returns a new Handler instance with a given u2ftoken.Token.
// Token cannot be nil.
func NewHandler(token Token) (*Handler, error) {
	if token == nil {
		return nil, errors.New("token is nil")
	}

	return &Handler{
		token: token,
		state: &u2fHIDState{
			sessions: map[uint32]*session{},
		},
	}, nil
}

// u2fPacket is implemented by U2F HID packets, and exposes methods that must be implemented
// to retrieve channel id, command, length, packet count and so on.
type u2fPacket interface {
	Channel() uint32
	ChannelBytes() [4]byte
	Command() uint8
	Length() uint16
	Count() uint16
}

// session holds informations about a single operation currently happening (MSG, PING...).
type session struct {
	data           []byte
	command        u2fHIDCommand
	total          uint64
	leftToRead     uint64
	lastSequence   uint8
	packetZeroSeen bool
}

// clear clears a session, setting everything to their default values.
func (s *session) clear() {
	s.data = nil
	s.command = 0
	s.total = 0
	s.leftToRead = 0
	s.lastSequence = 0
	s.packetZeroSeen = false
}

// u2fHIDState holds the global state of the U2FHID token, keeping track of whether it is still accumulating messages,
// all the outbound messages, all the sessions.
type u2fHIDState struct {
	outboundMsgs      [][]byte
	lastOutboundIndex int
	accumulatingMsgs  bool
	sessions          map[uint32]*session
	lastChannelID     uint32
}

// clear deletes the last channel id session, and sets outbound messages and its index, and channel id to zero.
func (u *u2fHIDState) clear() {
	u.sessions[u.lastChannelID].clear()
	u.outboundMsgs = nil
	u.lastOutboundIndex = 0
	u.lastChannelID = 0
	u.accumulatingMsgs = false
}
