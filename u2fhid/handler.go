package u2fhid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/gsora/fidati/leds"
)

var (
	state = &u2fHIDState{
		sessions: map[uint32]*session{},
	}
)

// zeroPad pads b with as many zeroes as needed to have len(b) == 64.
func zeroPad(b []byte) []byte {
	if len(b) == 64 {
		return b
	}

	for i := len(b); i < 64; i++ {
		b = append(b, 0)
	}

	return b
}

// Tx handles USB endpoint data outtake.
// res will always not be nil.
func (*Handler) Tx(buf []byte, lastErr error) (res []byte, err error) {
	if state.outboundMsgs == nil || state.accumulatingMsgs {
		return
	}

	if state.lastOutboundIndex == 0 {
		log.Println("found", len(state.outboundMsgs), "outbound messages")
	}

	if len(state.outboundMsgs) == 1 {
		b := &bytes.Buffer{}
		binary.Write(b, binary.LittleEndian, zeroPad(state.outboundMsgs[state.lastOutboundIndex]))

		res = b.Bytes()
		log.Println(res)
		log.Println("finished processing messages, clearing buffers")
		state.clear()
		return
	}

	if state.lastOutboundIndex == len(state.outboundMsgs) {
		log.Println("finished processing messages, clearing buffers")
		state.clear()
		return
	}

	b := &bytes.Buffer{}
	binary.Write(b, binary.LittleEndian, zeroPad(state.outboundMsgs[state.lastOutboundIndex]))
	state.lastOutboundIndex++

	res = b.Bytes()

	log.Println("processed message", state.lastOutboundIndex)

	return
}

// Rx handles data intake, parses messages and builds responses.
// res will always be nil.
func (*Handler) Rx(buf []byte, lastErr error) (res []byte, err error) {
	if buf == nil {
		return
	}

	msgs, parseErr := parseMsg(buf)
	if parseErr != nil {
		log.Println(parseErr)
		state.clear()
		return
	}

	state.outboundMsgs = msgs

	return
}

// parseMsg parses msg and constructs a slice of messages ready to be sent over the wire.
// Each response message is exactly 64 bytes in length.
func parseMsg(msg []byte) ([][]byte, error) {
	if len(msg) != 64 { // something's wrong
		return nil, fmt.Errorf("wrong message length, expected 64 but got %d", len(msg))
	}

	cmd := msg[4]
	isInit := isInitPkt(cmd)

	log.Println("msg ", msg)

	if isInit {
		log.Println("found init packet")
		ip := parseInitPkt(msg)

		s, ok := state.sessions[ip.Channel()]
		if !ok {
			s = &session{}
		}

		log.Println("command:", ip.Cmd.String())

		s.command = ip.Cmd
		s.total = uint64(ip.PayloadLength)
		s.data = make([]byte, 0, s.total)
		s.data = append(s.data, ip.Data...)
		s.leftToRead = uint64(int(ip.PayloadLength) - len(s.data))

		state.sessions[ip.Channel()] = s

		if s.total <= initPacketDataLen {
			// handle everything as a single entity
			return packetBuilder(s, ip)
		}

		state.accumulatingMsgs = true
	} else {
		log.Println("found continuation packet")
		cp := parseContinuationPkt(msg)

		session, ok := state.sessions[cp.Channel()]
		if !ok {
			return nil, fmt.Errorf("new continuation packet with id 0x%X, which was not seen before", cp.ChannelID)
		}

		if cp.SequenceNumber != 0 && cp.SequenceNumber-session.lastSequence != 1 {
			return nil, fmt.Errorf("found a continuation packet with non-sequential sequence number, expected %d but found %d", cp.SequenceNumber+1, session.lastSequence)
		}

		session.lastSequence = cp.SequenceNumber

		lastSize := len(session.data)
		// TODO: here we should count how many zeroes we should include in cp.Data, because some of them
		// are used in the U2F protocol.

		if session.leftToRead < continuationPacketDataLen {
			session.data = append(session.data, cp.Data[:session.leftToRead]...)
			session.leftToRead = 0
		} else {
			session.data = append(session.data, cp.Data...)
			session.leftToRead -= uint64(len(cp.Data))
		}

		log.Printf("read new %d bytes, last size %d, new size %d, total expected size %d", len(cp.Data), lastSize, len(session.data), session.total)

		if uint64(len(session.data)) > session.total {
			return nil, fmt.Errorf("read %d bytes while expecting %d", len(session.data), session.total)
		}

		if uint64(len(session.data)) != session.total {
			return nil, nil // we still need more data
		}

		log.Printf("finished reading data for channel 0x%X, total bytes %d", cp.Channel(), len(session.data))
		return packetBuilder(session, cp)
	}

	return nil, nil
}

// numPackets returns the number of packets needed to properly respond to a message.
func numPackets(rawMsgLen int) int {
	// 59 is the number of data bytes available in a continuation packet
	// 64 - (4 bytes channel id + 1 byte sequence number)
	return int(math.Ceil(float64(rawMsgLen) / float64(continuationPacketDataLen)))
}

// split splits msg into 64 bytes units.
func split(sizeFirst int, sizeRest int, msg []byte) [][]byte {
	numPktsNoInitial := numPackets(len(msg) - 57) // we exclude a packet, which will be built separately

	pktAmount := numPktsNoInitial + 1
	ret := make([][]byte, 0, pktAmount)

	lastIndex := 0

	for i := 0; i < pktAmount; i++ {
		if i == 0 {
			ret = append(ret, msg[0:sizeFirst])
			lastIndex = sizeFirst
			continue
		}

		if i+1 == pktAmount {
			ret = append(ret, msg[lastIndex:])
			return ret
		}

		ret = append(ret, msg[lastIndex:lastIndex+sizeRest])
		lastIndex = lastIndex + sizeRest
	}

	return ret
}

// broadcastReq responds to broadcast messages, sent with channel id [255, 255, 255, 255].
func broadcastReq(ip initPacket) []byte {
	if ip.Cmd != cmdInit {
		leds.Panic(fmt.Sprintf("found message for broadcast chan but command was %d instead of U2FHID_INIT", ip.Command()))
	}

	log.Println("found cmdInit on broadcast channel")

	b := new(bytes.Buffer)
	u := initResponse{
		standardResponse: standardResponse{
			Command:   ip.Command(),
			ChannelID: ip.ChannelID,
		},
		ProtocolVersion:    12,
		MajorDeviceVersion: 4,
		MinorDeviceVersion: 2,
		BuildDeviceVersion: 0,
		Capabilities:       0,
	}

	copy(u.Nonce[:], ip.Data)

	binary.BigEndian.PutUint16(u.Count[:], 17)
	err := binary.Write(b, binary.LittleEndian, u)
	if err != nil {
		leds.Panic(fmt.Sprintf("cannot serialize initResponse: %s", err.Error()))
	}
	return b.Bytes()
}

// packetBuilder builds response packages for a given session, depending on session.command.
func packetBuilder(session *session, pkt u2fPacket) ([][]byte, error) {
	log.Println("message", u2fHIDCommand(pkt.Command()))
	switch session.command {
	case cmdInit:
		ip, ok := pkt.(initPacket)
		if !ok {
			return nil, fmt.Errorf("found cmdInit packet, but said packet cannot be read as one")
		}

		if ip.Channel() != broadcastChan {
			return nil, fmt.Errorf("found a cmdInit, but not on the broadcast channel")
		}

		state.lastChannelID = broadcastChan

		return [][]byte{
			broadcastReq(ip),
		}, nil
	case cmdPing:
		pkts, err := handlePing(session, pkt)
		if err != nil {
			return nil, fmt.Errorf("error while handling ping, %w", err)
		}

		state.accumulatingMsgs = false
		state.lastChannelID = pkt.Channel()
		return pkts, nil
	case cmdMsg:
		pkts, err := handleMsg(session, pkt)
		if err != nil {
			return nil, fmt.Errorf("error while handling msg, %w", err)
		}

		state.accumulatingMsgs = false
		state.lastChannelID = pkt.Channel()
		return pkts, nil
	default:
		log.Printf("command %d not found, sending error payload", session.command)
		return generateError(invalidCmd, session, pkt), nil
	}
}
