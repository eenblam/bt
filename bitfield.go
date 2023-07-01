package bt

import (
	"errors"
	"fmt"
)

// "Bitfield" already used as an enum
type BField struct {
	// length may not be a multiple of 8!
	length int
	bs     []byte
}

// NewEmptyBitfield creates an all-zero bitfield from a length.
//
// This is almost always what you want to use when initiating a new downloader.
func NewEmptyBitfield(length int) (*BField, error) {
	if length < 0 {
		return nil, fmt.Errorf("unexpected negative length %d", length)
	}
	if length == 0 {
		return nil, errors.New("cannot create empty bitfield")
	}
	byteLength := length / 8
	// If there are remaining bits, add a byte to accommodate them.
	if length%8 > 0 {
		byteLength++
	}
	return &BField{
		bs:     make([]byte, byteLength),
		length: length,
	}, nil
}

// NewBitfield ensures a well-behaved BField is created at runtime.
//
// This eliminates certain classes of errors that are possible when creating a BField literal.
func NewBitfield(bs []byte, length int) (*BField, error) {
	if length < 0 {
		return nil, fmt.Errorf("unexpected negative length %d", length)
	}
	if length == 0 {
		return nil, errors.New("cannot create empty bitfield")
	}
	rangeUpper := len(bs) * 8
	if !((rangeUpper-8 <= length) && (length <= rangeUpper)) { // Out of range
		return nil, fmt.Errorf("bitfield length %d not in final byte of slice of length %d", length, len(bs))
	}
	return &BField{
		bs:     bs,
		length: length,
	}, nil
}

// Get value of i-th bit, indexed from 0 to (LENGTH-1), big-endian.
//
// Errors if i exceeds bitfield length or (if b.length is too large) byte slice length.
func (b *BField) Get(i int) (bool, error) {
	if b.length <= i { // index bits from 0
		return false, fmt.Errorf("bitfield has %d bytes and %d bits, got index %d", len(b.bs), b.length, i)
	}
	byteIndex := i / 8
	bitIndex := 7 - (i % 8) // index 0-7 from high bit to low bit
	if byteIndex >= len(b.bs) {
		return false, fmt.Errorf("bitfield has %d bytes, with bit length %d. Got index into byte %d (BField.Length is misconfigured!)",
			len(b.bs), b.length, byteIndex)
	}
	// Just get the bit, see if result is 0
	masked := b.bs[byteIndex] & (1 << bitIndex)
	return masked != 0, nil
}

// Set value of i-th bit to bl, indexed from 0 to (LENGTH-1), big-endian.
func (b *BField) Set(i int, bl bool) error {
	if b.length <= i { // index bits from 0
		return fmt.Errorf("bitfield has %d bytes and %d bits, got index %d", len(b.bs), b.length, i)
	}
	byteIndex := i / 8
	bitIndex := 7 - (i % 8) // index 0-7 from high bit to low bit
	if byteIndex >= len(b.bs) {
		return fmt.Errorf("bitfield has %d bytes, with bit length %d. Got index into byte %d (BField.Length is misconfigured!)",
			len(b.bs), b.length, byteIndex)
	}
	if bl { // set true bit via OR
		b.bs[byteIndex] = b.bs[byteIndex] | (1 << bitIndex)
	} else { // set false bit via AND NOT
		b.bs[byteIndex] = b.bs[byteIndex] & ^(1 << bitIndex)
	}
	return nil
}

// Swap value of i-th bit, indexed from 0 to (LENGTH-1), big-endian.
func (b *BField) Swap(i int) error {
	if b.length <= i { // index bits from 0
		return fmt.Errorf("bitfield has %d bytes and %d bits, got index %d", len(b.bs), b.length, i)
	}
	byteIndex := i / 8
	bitIndex := 7 - (i % 8) // index 0-7 from high bit to low bit
	if byteIndex >= len(b.bs) {
		return fmt.Errorf("bitfield has %d bytes, with bit length %d. Got index into byte %d (BField.Length is misconfigured!)",
			len(b.bs), b.length, byteIndex)
	}
	masked := b.bs[byteIndex] & (1 << bitIndex)
	if masked == 0 { // false. set true bit via OR
		b.bs[byteIndex] = b.bs[byteIndex] | (1 << bitIndex)
	} else { // set false bit via AND NOT
		b.bs[byteIndex] = b.bs[byteIndex] & ^(1 << bitIndex)
	}
	return nil
}

func (b *BField) Length() int {
	return b.length
}
