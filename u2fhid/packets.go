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

// base size = 5, data size = 64 - 5 = 59
type continuationPacket struct {
	ChannelID      [4]byte
	SequenceNumber uint8
	Data           []byte
}

func (i continuationPacket) ChannelBytes() [4]byte {
	return i.ChannelID
}

func (i continuationPacket) Command() uint8 {
	return 0
}

func (i continuationPacket) Length() uint16 {
	return 0
}

func (i continuationPacket) Count() uint16 {
	return uint16(i.SequenceNumber)
}

func (i continuationPacket) Channel() uint32 {
	return binary.BigEndian.Uint32(i.ChannelID[:])
}

func (i continuationPacket) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("channel id: %v, ", i.ChannelID))
	s.WriteString(fmt.Sprintf("sequence number: %d, ", i.SequenceNumber))
	s.WriteString(fmt.Sprintf("data: %v", i.Data))
	return s.String()
}

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

// FIDO U2F HID Protocol Specification, pg 4, "2.4 Message- and packet structure"
func parseContinuationPkt(msg []byte) continuationPacket {
	i := continuationPacket{
		SequenceNumber: msg[4],
	}

	copy(i.ChannelID[:], msg[0:4])

	i.Data = msg[5:]

	return i
}

type initPacket struct {
	ChannelID     [4]byte
	Cmd           u2fHIDCommand
	PayloadLength uint16
	Data          []byte
}

func (i initPacket) ChannelBytes() [4]byte {
	return i.ChannelID
}

func (i initPacket) Command() uint8 {
	return uint8(i.Cmd)
}

func (i initPacket) Length() uint16 {
	return i.PayloadLength
}

func (i initPacket) Count() uint16 {
	return 0
}

func (i initPacket) Channel() uint32 {
	return binary.BigEndian.Uint32(i.ChannelID[:])
}

func (i initPacket) String() string {
	s := strings.Builder{}
	s.WriteString(fmt.Sprintf("channel id: %v, ", i.ChannelID))
	s.WriteString(fmt.Sprintf("command: %d, ", i.Command()))
	s.WriteString(fmt.Sprintf("payload length: %d, ", i.PayloadLength))
	s.WriteString(fmt.Sprintf("data: %v", i.Data))
	return s.String()
}

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

// FIDO U2F HID Protocol Specification, pg 4, "2.4 Message- and packet structure"
// A packet is an Init one if bit 7 is set.
func isInitPkt(cmd uint8) bool {
	return cmd>>7 == 1
}

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
