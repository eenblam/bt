package bt

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

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

func ParseMessage(r io.Reader) (MType, []byte, error) {
	return Choke, nil, nil
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
