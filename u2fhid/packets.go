package u2fhid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	initPacketDataLen         = 57
	continuationPacketDataLen = 59
)

// continuationPacket is a U2FHID message packet for which the command byte has the seventh bit not set.
// Base size = 5, data size = 64 - 5 = 59.
type continuationPacket struct {
	ChannelID      [4]byte
	SequenceNumber uint8
	Data           []byte
}

// ChannelBytes implements u2fPacket interface.
func (i continuationPacket) ChannelBytes() [4]byte {
	return i.ChannelID
}

// Command implements u2fPacket interface.
func (i continuationPacket) Command() uint8 {
	return 0
}

// Length implements u2fPacket interface.
func (i continuationPacket) Length() uint16 {
	return 0
}

// Count implements u2fPacket interface.
func (i continuationPacket) Count() uint16 {
	return uint16(i.SequenceNumber)
}

// Channel implements u2fPacket interface.
func (i continuationPacket) Channel() uint32 {
	return binary.BigEndian.Uint32(i.ChannelID[:])
}

// String implements fmt.Stringer interface.
func (i continuationPacket) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("channel id: %v, ", i.ChannelID))
	s.WriteString(fmt.Sprintf("sequence number: %d, ", i.SequenceNumber))
	s.WriteString(fmt.Sprintf("data: %v", i.Data))
	return s.String()
}

// Bytes returns the byte slice representation of a continuation packet.
func (i continuationPacket) Bytes() []byte {
	s := struct {
		ChannelID      [4]byte
		SequenceNumber uint8
	}{
		i.ChannelID,
		i.SequenceNumber,
	}

	b := new(bytes.Buffer)
	err := binary.Write(b, binary.BigEndian, s)
	if err != nil {
		panic(fmt.Sprintf("cannot format continuationPacket, %s", err.Error()))
	}

	return append(b.Bytes(), i.Data...)
}

// parseContinuationPkt parses msg as a continuation packet, and returns a continuationPacket instance.
// FIDO U2F HID Protocol Specification, pg 4, "2.4 Message- and packet structure"
func parseContinuationPkt(msg []byte) continuationPacket {
	i := continuationPacket{
		SequenceNumber: msg[4],
	}

	copy(i.ChannelID[:], msg[0:4])

	i.Data = msg[5:]

	return i
}

// initPacket is a U2FHID message packet for which the command byte has the seventh bit set.
type initPacket struct {
	ChannelID     [4]byte
	Cmd           u2fHIDCommand
	PayloadLength uint16
	Data          []byte
}

// ChannelBytes implements u2fPacket interface.
func (i initPacket) ChannelBytes() [4]byte {
	return i.ChannelID
}

// Command implements u2fPacket interface.
func (i initPacket) Command() uint8 {
	return uint8(i.Cmd)
}

// Length implements u2fPacket interface.
func (i initPacket) Length() uint16 {
	return i.PayloadLength
}

// Count implements u2fPacket interface.
func (i initPacket) Count() uint16 {
	return 0
}

// Channel implements u2fPacket interface.
func (i initPacket) Channel() uint32 {
	return binary.BigEndian.Uint32(i.ChannelID[:])
}

// String implements fmt.Stringer interface.
func (i initPacket) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("channel id: %v, ", i.ChannelID))
	s.WriteString(fmt.Sprintf("command: %d, ", i.Command()))
	s.WriteString(fmt.Sprintf("payload length: %d, ", i.PayloadLength))
	s.WriteString(fmt.Sprintf("data: %v", i.Data))
	return s.String()
}

// parseInitPkt parses msg as a init packet, and returns a initPacket instance.
// FIDO U2F HID Protocol Specification, pg 4, "2.4 Message- and packet structure"
func parseInitPkt(msg []byte) initPacket {
	i := initPacket{
		Cmd: u2fHIDCommand(msg[4]),
	}

	copy(i.ChannelID[:], msg[0:4])

	i.PayloadLength = binary.BigEndian.Uint16([]byte{msg[5], msg[6]})

	i.Data = msg[7:]

	return i
}

// isInitPkt returns true if the seventh bit of cmd is set.
// FIDO U2F HID Protocol Specification, pg 4, "2.4 Message- and packet structure"
func isInitPkt(cmd uint8) bool {
	return cmd>>7 == 1
}

// genPackets generates response packets for msg, command cmd and channel id chanID.
func genPackets(msg []byte, cmd u2fHIDCommand, chanID [4]byte) ([][]byte, error) {
	numPktsNoInitial := numPackets(len(msg) - 64) // we exclude a packet, which will be built separately

	ret := make([][]byte, 0, numPktsNoInitial+1)

	ulog.Println("expected number of packets:", numPktsNoInitial+1)

	sequence := 0
	for i, packetPayload := range split(initPacketDataLen, continuationPacketDataLen, msg) {
		if i == 0 {
			b := new(bytes.Buffer)
			u := standardResponse{
				Command:   uint8(cmd),
				ChannelID: chanID,
			}

			binary.BigEndian.PutUint16(u.Count[:], uint16(len(msg)))
			ulog.Println("length", uint16(len(msg)), "bytes", u.Count)
			err := binary.Write(b, binary.LittleEndian, u)
			if err != nil {
				return nil, fmt.Errorf("cannot serialize msg payload, %w", err)
			}

			initPingMsg := append(b.Bytes(), packetPayload...)
			ret = append(ret, initPingMsg)

			ulog.Println("built packet", i)
			continue
		}

		var cc continuationPacket
		cc.SequenceNumber = uint8(sequence)
		cc.ChannelID = chanID
		cc.Data = packetPayload
		ret = append(ret, cc.Bytes())
		ulog.Println("built packet", i)
		sequence++
	}

	return ret, nil
}
