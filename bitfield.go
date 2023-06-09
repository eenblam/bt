package bt

import (
	"bytes"
	"errors"
	"fmt"
)

const nextUnknown = -1

// "Bitfield" already used as an enum
type BField struct {
	// length may not be a multiple of 8!
	length int
	bs     []byte
	// Starts at nextUnknown: don't know state of first bit
	// Set to b.length when last bit has been set
	// May also be updated if set/swap changes a value *below* this value
	// e.g. if nextFalse is 10, Set(8, false) will update nextFalse to 8
	nextFalse int
	nextTrue  int
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
		bs:        make([]byte, byteLength),
		length:    length,
		nextFalse: -1,
		nextTrue:  -1,
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
		bs:        bs,
		length:    length,
		nextFalse: -1,
		nextTrue:  -1,
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
		if i < b.nextTrue {
			b.nextTrue = i
		}
	} else { // set false bit via AND NOT
		b.bs[byteIndex] = b.bs[byteIndex] & ^(1 << bitIndex)
		if i < b.nextFalse {
			b.nextFalse = i
		}
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
		if i < b.nextTrue {
			b.nextTrue = i
		}
	} else { // set false bit via AND NOT
		b.bs[byteIndex] = b.bs[byteIndex] & ^(1 << bitIndex)
		if i < b.nextFalse {
			b.nextFalse = i
		}
	}
	return nil
}

func (b *BField) Length() int {
	return b.length
}

// Searches for next 0 bit in BField, returns (index, done).
//
// When exhausted, done will equal BField.Length().
func (b *BField) NextFalse() (int, bool) {
	// Implementation note: these *could* be joined into something DRY,
	// but you'd have to do some juggling of the two different
	// internal state variables b.nextFalse and b.nextTrue.
	// Separate methods prevents related bugs.
	if b.nextFalse >= b.length { // All true
		return b.nextFalse, true
	}
	var currentByte, currentBit, maxBytes int
	if b.nextFalse == nextUnknown { // Have to start fresh
		b.nextFalse = 0
		currentByte, currentBit = 0, 0
	} else { // Start from current nextFalse
		currentByte = b.nextFalse / 8
		currentBit = b.nextFalse % 8
	}
	maxBytes = len(b.bs)
	for ; currentByte < maxBytes; currentByte++ {
		for ; currentBit < 8; currentBit++ {
			b.nextFalse = (currentByte * 8) + currentBit
			if b.nextFalse >= b.length { // don't run into extra bits beyond length
				return b.nextFalse, true
			}
			if (b.bs[currentByte] & (1 << (7 - currentBit))) == 0 {
				return b.nextFalse, false
			}
		}
		// After first use, stop searching from original currentBit
		currentBit = 0
	}
	// Exhausted bytes with nothing found!
	// We should've caught this prior, but just in case (and to appease compiler)
	b.nextFalse = b.length
	return b.nextFalse, true
}

// Searches for next 1 bit in BField, returns (index, done).
//
// When exhausted, done will equal BField.Length().
func (b *BField) NextTrue() (int, bool) {
	if b.nextTrue >= b.length { // All true
		return b.nextTrue, true
	}
	var currentByte, currentBit, maxBytes int
	if b.nextTrue == nextUnknown { // Have to start fresh
		b.nextTrue = 0
		currentByte, currentBit = 0, 0
	} else { // Start from current nextTrue
		currentByte = b.nextTrue / 8
		currentBit = b.nextTrue % 8
	}
	maxBytes = len(b.bs)
	for ; currentByte < maxBytes; currentByte++ {
		for ; currentBit < 8; currentBit++ {
			b.nextTrue = (currentByte * 8) + currentBit
			if b.nextTrue >= b.length { // don't run into extra bits beyond length
				return b.nextTrue, true
			}
			if (b.bs[currentByte] & (1 << (7 - currentBit))) != 0 {
				return b.nextTrue, false
			}
		}
		// After first use, stop searching from original currentBit
		currentBit = 0
	}
	// Exhausted bytes with nothing found!
	// We should've caught this prior, but just in case (and to appease compiler)
	b.nextTrue = b.length
	return b.nextTrue, true
}

// Sub takes the difference of b and a: the bits in b that are not in a.
// More precisely, (b & (^a)). Errors on length mismatches.
//
// Use this to find the pieces that a peer has that you don't:
// wantPieces := peerPieces.Sub(downloaderPieces)
func (b *BField) Sub(a *BField) (*BField, error) {
	if b.length != a.length {
		return nil, fmt.Errorf("bit lengths unequal - expected %d, got %d", b.length, a.length)
	}
	if len(b.bs) != len(a.bs) { // avoid OOB read on manually created BFields
		return nil, fmt.Errorf("byte array lengths unequal - expected %d, got %d", len(b.bs), len(a.bs))
	}
	out := make([]byte, len(b.bs))
	for i, x := range b.bs {
		out[i] = x & (^a.bs[i])
	}
	// set each iterator state to minimum of the two inputs
	var nextFalse, nextTrue int
	if b.nextFalse <= a.nextFalse {
		nextFalse = b.nextFalse
	} else {
		nextFalse = a.nextFalse
	}
	if b.nextTrue <= a.nextTrue {
		nextTrue = b.nextTrue
	} else {
		nextTrue = a.nextTrue
	}
	return &BField{
		bs:        out,
		length:    b.length,
		nextFalse: nextFalse,
		nextTrue:  nextTrue,
	}, nil
}

// SubInto takes the difference of b and a (the bits in b that are not in a)
// and writes them to "into".
// More precisely, into = (b & (^a)). Errors on length mismatches.
//
// Use this to find the pieces that a peer has that you don't:
// peerPieces.Sub(piecesToRequest, downloaderPieces)
func (b *BField) SubInto(into, a *BField) error {
	if b.length != into.length {
		return fmt.Errorf("\"into\" bit length doesn't match source - expected %d, got %d", b.length, into.length)
	}
	if b.length != a.length {
		return fmt.Errorf("\"a\" bit length doesn't match source - expected %d, got %d", b.length, a.length)
	}
	if len(b.bs) != len(into.bs) { // avoid OOB read on manually created BFields
		return fmt.Errorf("\"into\" byte array length doesn't match source - expected %d, got %d", len(b.bs), len(into.bs))
	}
	if len(b.bs) != len(a.bs) {
		return fmt.Errorf("\"a\" byte array length doesn't match source - expected %d, got %d", len(b.bs), len(a.bs))
	}
	for i, x := range b.bs {
		into.bs[i] = x & (^a.bs[i])
	}
	// set each iterator state to minimum of the two inputs
	if b.nextFalse <= a.nextFalse {
		into.nextFalse = b.nextFalse
	} else {
		into.nextFalse = a.nextFalse
	}
	if b.nextTrue <= a.nextTrue {
		into.nextTrue = b.nextTrue
	} else {
		into.nextTrue = a.nextTrue
	}
	return nil
}

// Equal compares two BFields for equality, returning an error for unequal lengths.
func (b *BField) Equal(a *BField) (bool, error) {
	if b.length != a.length {
		return false, fmt.Errorf("bit lengths unequal - expected %d, got %d", b.length, a.length)
	}
	if len(b.bs) != len(a.bs) { // avoid OOB read on manually created BFields
		return false, fmt.Errorf("byte array lengths unequal - expected %d, got %d", len(b.bs), len(a.bs))
	}
	return bytes.Equal(b.bs, a.bs), nil
}
