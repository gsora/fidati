package u2fhid

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// u2fError represent some kind of error happened during the u2fhid processing.
//go:generate stringer -type=u2fError
type u2fError uint8

const (
	// no error
	none u2fError = 0

	// invalid command
	invalidCmd u2fError = 1

	// invalid parameter
	invalidPar u2fError = 2

	// invalid length
	invalidLen u2fError = 3

	// invalid sequence number
	invalidSeq u2fError = 4

	// timeout
	msgTimeout u2fError = 5

	// communication channel is busy
	channelBusy u2fError = 6

	// a channel lock is required
	lockRequired u2fError = 10

	// invalid channel id
	invalidCid u2fError = 11

	// other kind of error
	other u2fError = 127
)

// generateError generates a u2fError payload ready to be sent on the wire.
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
