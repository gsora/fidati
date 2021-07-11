package u2fhid

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/gsora/fidati/internal/flog"
)

// zeroPad pads b with as many zeroes as needed to have len(b) == 64.
func zeroPad(b []byte) []byte {
	if len(b) == 64 {
		return b
	}

	nb := make([]byte, 64)
	copy(nb, b)

	return nb
}

// Tx handles USB endpoint data outtake.
// res will always not be nil.
func (h *Handler) Tx(buf []byte, lastErr error) (res []byte, err error) {
	if h.state.outboundMsgs == nil || h.state.accumulatingMsgs {
		return
	}

	if h.state.lastOutboundIndex == 0 {
		flog.Logger.Println("found", len(h.state.outboundMsgs), "outbound messages")
	}

	if len(h.state.outboundMsgs) == 1 {
		b := &bytes.Buffer{}
		binary.Write(b, binary.LittleEndian, zeroPad(h.state.outboundMsgs[h.state.lastOutboundIndex]))

		res = b.Bytes()
		flog.Logger.Println(res)
		flog.Logger.Println("finished processing messages, clearing buffers")
		h.state.clear()
		return
	}

	if h.state.lastOutboundIndex == len(h.state.outboundMsgs) {
		flog.Logger.Println("finished processing messages, clearing buffers")
		h.state.clear()
		return
	}

	b := &bytes.Buffer{}
	binary.Write(b, binary.LittleEndian, zeroPad(h.state.outboundMsgs[h.state.lastOutboundIndex]))
	h.state.lastOutboundIndex++

	res = b.Bytes()

	flog.Logger.Println("processed message", h.state.lastOutboundIndex)

	return
}

// Rx handles data intake, parses messages and builds responses.
// res will always be nil.
func (h *Handler) Rx(buf []byte, lastErr error) (res []byte, err error) {
	if buf == nil {
		return
	}

	msgs, err := h.parseMsg(buf)
	if err != nil {
		flog.Logger.Println(err)
		h.state.clear()
		return
	}

	h.state.outboundMsgs = msgs

	return
}

// parseMsg parses msg and constructs a slice of messages ready to be sent over the wire.
// Each response message is exactly 64 bytes in length.
func (h *Handler) parseMsg(msg []byte) ([][]byte, error) {
	if len(msg) != 64 { // something's wrong
		return nil, fmt.Errorf("wrong message length, expected 64 but got %d", len(msg))
	}

	cmd := msg[4]
	isInit := isInitPkt(cmd)

	flog.Logger.Println("msg ", msg)

	if isInit {
		return h.handleInitPacket(msg)
	} else {
		return h.handleContinuationPacket(msg)
	}
}

// handleContinuationPacket handles parsing and state update for continuation packets.
func (h *Handler) handleContinuationPacket(msg []byte) ([][]byte, error) {
	flog.Logger.Println("found continuation packet")
	cp := parseContinuationPkt(msg)

	session, ok := h.state.sessions[cp.Channel()]
	if !ok {
		return nil, fmt.Errorf("new continuation packet with id 0x%X, which was not seen before", cp.ChannelID)
	}

	var seqError bool
	switch {
	case cp.SequenceNumber == 0:
		if session.packetZeroSeen {
			seqError = true
		}
	case cp.SequenceNumber-session.lastSequence != 1:
		seqError = true
	}

	if seqError {
		return nil, fmt.Errorf("found a continuation packet with non-sequential sequence number, expected %d but found %d", cp.SequenceNumber+1, session.lastSequence)
	}

	session.packetZeroSeen = true
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

	flog.Logger.Printf("read new %d bytes, last size %d, new size %d, total expected size %d", len(cp.Data), lastSize, len(session.data), session.total)

	if len(session.data) != int(session.total) {
		return nil, nil // we still need more data
	}

	flog.Logger.Printf("finished reading data for channel 0x%X, total bytes %d", cp.Channel(), len(session.data))
	return h.packetBuilder(session, cp)
}

// handleInitPacket handles parsing and state setup for initialization packets.
func (h *Handler) handleInitPacket(msg []byte) ([][]byte, error) {
	flog.Logger.Println("found init packet")
	ip := parseInitPkt(msg)

	s, ok := h.state.sessions[ip.Channel()]
	if !ok {
		s = &session{}
	}

	flog.Logger.Println("command:", ip.Cmd.String())

	s.command = ip.Cmd
	s.total = uint64(ip.PayloadLength)
	s.data = make([]byte, 0, s.total)
	s.data = append(s.data, ip.Data...)
	s.leftToRead = uint64(int(ip.PayloadLength) - len(s.data))

	h.state.sessions[ip.Channel()] = s

	if s.total <= initPacketDataLen {
		// handle everything as a single entity
		return h.packetBuilder(s, ip)
	}

	h.state.accumulatingMsgs = true

	return nil, nil
}

// numPackets returns the number of packets needed to properly respond to a message.
func numPackets(rawMsgLen int) int {
	// 59 is the number of data bytes available in a continuation packet
	// 64 - (4 bytes channel id + 1 byte sequence number)
	return int(math.Ceil(float64(rawMsgLen) / float64(continuationPacketDataLen)))
}

// broadcastReq responds to broadcast messages, sent with channel id [255, 255, 255, 255].
func broadcastReq(ip initPacket) ([]byte, error) {
	if ip.Cmd != cmdInit {
		return nil, fmt.Errorf("found message for broadcast chan but command was %d instead of U2FHID_INIT", ip.Command())
	}

	flog.Logger.Println("found cmdInit on broadcast channel")

	assignedChannelID := make([]byte, 4)
	_, err := rand.Read(assignedChannelID)
	if err != nil {
		return nil, fmt.Errorf("cannot generate random channel ID, %w", err)
	}

	flog.Logger.Println("created new channel id", assignedChannelID)

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
	copy(u.AssignedChannelID[:], assignedChannelID)

	binary.BigEndian.PutUint16(u.Count[:], 17)
	err = binary.Write(b, binary.LittleEndian, u)
	if err != nil {
		return nil, fmt.Errorf("cannot serialize initResponse: %w", err)
	}

	flog.Logger.Println("finished broadcastReq")
	return b.Bytes(), nil
}

// packetBuilder builds response packages for a given session, depending on session.command.
func (h *Handler) packetBuilder(session *session, pkt u2fPacket) ([][]byte, error) {
	flog.Logger.Println("message", u2fHIDCommand(pkt.Command()))

	if ch, handled := h.commandMappings[session.command]; handled {
		flog.Logger.Println("found command to be handled via command mappings:", session.command)

		pkts, err := genPackets(
			ch(session.data[:session.total], pkt.Channel()),
			session.command,
			pkt.ChannelBytes(),
		)

		if err != nil {
			return nil, fmt.Errorf("error while handling msg, %w", err)
		}

		h.state.accumulatingMsgs = false
		h.state.lastChannelID = pkt.Channel()
		return pkts, nil
	}

	// use standard u2fhid commands
	switch session.command {
	case cmdInit:
		ip, ok := pkt.(initPacket)
		if !ok {
			return nil, fmt.Errorf("found cmdInit packet, but said packet cannot be read as one")
		}

		if ip.Channel() != broadcastChan {
			return nil, fmt.Errorf("found a cmdInit, but not on the broadcast channel")
		}

		h.state.lastChannelID = broadcastChan
		h.state.accumulatingMsgs = false

		ret, err := broadcastReq(ip)
		if err != nil {
			return nil, err
		}

		return [][]byte{ret}, nil
	case cmdPing:
		pkts, err := handlePing(session, pkt)
		if err != nil {
			return nil, fmt.Errorf("error while handling ping, %w", err)
		}

		h.state.accumulatingMsgs = false
		h.state.lastChannelID = pkt.Channel()
		return pkts, nil
	case cmdMsg:
		pkts, err := h.handleMsg(session, pkt)
		if err != nil {
			return nil, fmt.Errorf("error while handling msg, %w", err)
		}

		h.state.accumulatingMsgs = false
		h.state.lastChannelID = pkt.Channel()
		return pkts, nil
	default:
		flog.Logger.Printf("command %d not found, sending error payload", session.command)
		return generateError(invalidCmd, pkt), nil
	}
}
