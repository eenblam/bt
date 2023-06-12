package bt

import (
	"bytes"
	"crypto/sha1"
	"io"
	"testing"
)

func TestExpect(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name            string
		InputReader     io.Reader
		InputExpected   []byte
		InputExpectNext []byte
		WantError       bool
	}{
		{
			Name:            "",
			InputReader:     bytes.NewReader([]byte(`1234567890`)),
			InputExpected:   []byte(`1234`),
			InputExpectNext: []byte(`567890`),
			WantError:       false,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			buf := make([]byte, len(c.InputExpected))
			err := Expect(c.InputReader, buf, c.InputExpected)
			if c.WantError {
				if err == nil {
					t.Fatal("Wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			// Confirm the next bytes are as expected
			err = Expect(c.InputReader, buf, c.InputExpectNext)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
		})
	}
}

// TestHandshakeSymmetry confirms that MakeHandshake and ParseHandshake are inverses.
func TestHandshakeSymmetry(t *testing.T) {
	// sha1.Sum() will give us a [20]byte
	infoHash := sha1.Sum([]byte("infohash"))
	peerId := sha1.Sum([]byte("peerid"))

	r := bytes.NewReader(MakeHandshake(infoHash, peerId))
	err := ParseHandshake(r, infoHash, peerId)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestParseMessage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name        string
		InputReader io.Reader
		WantType    MType
		WantPayload []byte
		WantError   bool
	}{
		// Leave 0xff after each to confirm we didn't over-read
		{
			Name:        "parses Choke",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 1, 0, 0xff}),
			WantType:    Choke,
			WantError:   false,
		},
		{
			Name:        "return error for bad Choke length",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 5, 0, 0xff, 0xff, 0xff, 0xff, 0xff}),
			WantError:   true,
		},
		{
			Name:        "parses Unchoke",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 1, 1, 0xff}),
			WantType:    Unchoke,
			WantError:   false,
		},
		{
			Name:        "parses Interested",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 1, 2, 0xff}),
			WantType:    Interested,
			WantError:   false,
		},
		{
			Name:        "parses NotInterested",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 1, 3, 0xff}),
			WantType:    NotInterested,
			WantError:   false,
		},
		{
			Name:        "parses Have",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 5, 4, 1, 2, 3, 4, 0xff}),
			WantType:    Have,
			WantPayload: []byte{1, 2, 3, 4},
			WantError:   false,
		},
		{
			Name: "parses Bitfield", // Variable length, can just treat this like Request for now
			InputReader: bytes.NewReader([]byte{0, 0, 0, 13, 5,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				0xff}),
			WantType:    Bitfield,
			WantPayload: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			WantError:   false,
		},
		{
			Name: "parses Request",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 13, 6,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				0xff}),
			WantType:    Request,
			WantPayload: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			WantError:   false,
		},
		{
			Name: "parses Piece", // Same as Request if length >= 9
			InputReader: bytes.NewReader([]byte{0, 0, 0, 13, 7,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				0xff}),
			WantType:    Piece,
			WantPayload: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			WantError:   false,
		},
		{
			Name: "fails to parse Piece if length < 9", // Same as Request if length >= 9
			InputReader: bytes.NewReader([]byte{0, 0, 0, 8, 7,
				1, 2, 3, 4, 5, 6, 7,
				0xff}),
			WantType:  Piece,
			WantError: true,
		},
		{
			Name: "parses Cancel", // same as Reques for this purpose
			InputReader: bytes.NewReader([]byte{0, 0, 0, 13, 8,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				0xff}),
			WantType:    Cancel,
			WantPayload: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			WantError:   false,
		},
		{
			Name:        "parses Port",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 3, 9, 1, 2, 0xff}),
			WantType:    Port,
			WantPayload: []byte{1, 2},
			WantError:   false,
		},
		{
			Name:        "parses KeepAlive",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 0, 0xff}),
			WantType:    KeepAlive,
			WantError:   false,
		},
		// Errors not specific to a message type
		{
			Name:        "fails if length too short",
			InputReader: bytes.NewReader([]byte{0, 0, 0}),
			WantError:   true,
		},
		{
			Name:        "fails if not KeepAlive and no id/type byte",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 1}),
			WantError:   true,
		},
		{
			Name:        "fails for unknown id/type byte",
			InputReader: bytes.NewReader([]byte{0, 0, 0, 1, 0xff}),
			WantError:   true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			m, err := ParseMessage(c.InputReader)
			if c.WantError {
				if err == nil {
					t.Fatal("Wanted error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			// Confirm type, payload, next byte
			if c.WantType != m.Type {
				t.Fatalf("want type %s, got %s", c.WantType, m.Type)
			}
			if !bytes.Equal(c.WantPayload, m.Payload) {
				t.Fatalf("\n\twant payload: %x\n\tgot payload:  %x", c.WantPayload, m.Payload)
			}
			// Confirm the next byte is as expected
			err = Expect(c.InputReader, []byte{0}, []byte{0xff})
			if err != nil {
				t.Fatalf("next byte from reader doesn't match: %s", err)
			}
		})
	}
}
