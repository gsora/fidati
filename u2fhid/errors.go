package u2fhid

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

//go:generate stringer -type=u2fError
type u2fError uint8

const (
	None         u2fError = 0
	InvalidCmd   u2fError = 1
	InvalidPar   u2fError = 2
	InvalidLen   u2fError = 3
	InvalidSeq   u2fError = 4
	MsgTimeout   u2fError = 5
	ChannelBusy  u2fError = 6
	LockRequired u2fError = 10
	InvalidCid   u2fError = 11
	Other        u2fError = 127
)

func generateError(code u2fError, session *session, pkt u2fPacket) [][]byte {
	b := new(bytes.Buffer)

	u := standardResponse{
		Command:   uint8(cmdError),
		ChannelID: pkt.ChannelBytes(),
	}

	binary.BigEndian.PutUint16(u.Count[:], uint16(1))
	err := binary.Write(b, binary.LittleEndian, u)
	if err != nil {
		panic(fmt.Errorf("cannot serialize msg payload, %w", err).Error())
	}

	final := append(b.Bytes(), uint8(code))

	return [][]byte{
		final,
	}
}
