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
