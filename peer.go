package bt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

//go:generate stringer -type=MType
type MType byte

const (
	Choke MType = iota
	Unchoke
	Interested
	NotInterested
	Have
	Bitfield
	Request
	Piece
	Cancel
	Port      // Newer, for DHT trackers
	KeepAlive // Not an actual message id, but when length=0000
)

var HandshakePrefix = []byte(`19BitTorrent protocol00000000`)

func MakeHandshake(infoHash, peerId [20]byte) []byte {
	out := make([]byte, 69) // 29 + 20 + 20
	copy(out, HandshakePrefix)
	copy(out[29:], infoHash[:])
	copy(out[49:], peerId[:])
	return out
}

// Parse 19BitTorrent protocol00000000<infoHash><peerId>
func ParseHandshake(r io.Reader, infoHash, peerId [20]byte) error {
	//e := NewStaticExpecter(29) // max(29, 20, 20)
	buf := make([]byte, 29) // max(29, 20, 20)
	return ExpectMany(r, buf, HandshakePrefix, infoHash[:], peerId[:])
}

func Expect(from io.Reader, buf, want []byte) error {
	if len(want) > cap(buf) {
		log.Printf("Expect: increasing buffer capacity from %d to %d", cap(buf), len(want))
		buf = make([]byte, len(want))
	}
	buf = buf[:len(want)]
	if _, err := io.ReadFull(from, buf); err != nil {
		return fmt.Errorf("Expect: couldn't read %x from reader: %s", want, err)
	}
	if !bytes.Equal(buf, want) {
		return fmt.Errorf("Expect: want %x, got %x", want, buf)
	}
	return nil
}

func ExpectMany(from io.Reader, buf []byte, wants ...[]byte) error {
	for _, w := range wants {
		if err := Expect(from, buf, w); err != nil {
			return err
		}
	}
	return nil
}

type Message struct {
	Type    MType
	Length  uint32 // int?
	Payload []byte
}

// ParseMessage tries to read a peer message from the reader, returning an error on failure
func ParseMessage(r io.Reader) (*Message, error) {
	// Parse length as BigEndian uint32
	var length uint32
	err := binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse message length: %s", err)
	}
	fmt.Println(length)
	// if 0000, it's keep-alive
	if length == 0 {
		return &Message{Type: KeepAlive}, nil
	}
	buf := make([]byte, length)
	// Parse message id/type (one byte)
	if _, err = io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("couldn't parse message id byte: %s", err)
	}
	mType := MType(buf[0])
	if length == 1 {
		return ValidateMessage(&Message{Type: mType, Length: length})
	} else {
		return ValidateMessage(&Message{Type: mType, Length: length, Payload: buf[1:]})
	}
}

// ValidateMessage ensures a message's length is appropriate for its type.
// Checks both computed Length field and actual length of parsed Payload.
func ValidateMessage(m *Message) (*Message, error) {
	want := m.Length
	switch m.Type {
	case Choke, Unchoke, Interested, NotInterested: // No payload
		want = 1
	case Have: // Fixed length with payload
		want = 5
	case Request, Cancel: // Fixed length 13 with payload: <index><begin><length>
		want = 13
	case Port: // Fixed length with payload
		// (Used in newer versions for DHT tracker)
		want = 3
	case Bitfield: // Variable length
		//TODO validate against length of pieces?
	case Piece: // Variable length: <index><begin><block...> (at least 9)
		if m.Length < 9 {
			return nil, fmt.Errorf("expected length >=9 for Piece message, got %d", m.Length)
		}
	default:
		return nil, fmt.Errorf("unknown Message ID/Type: %b", m.Type)
	}
	if m.Length != want {
		return nil, fmt.Errorf("expected length %d for %s message, got %d", want, m.Type, m.Length)
	}
	if len(m.Payload) != int(want)-1 {
		return nil, fmt.Errorf("expected payload of length %d for %s message, got %d", want, m.Type, len(m.Payload))
	}
	return m, nil
}
