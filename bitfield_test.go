package bt

import (
	"bytes"
	"testing"
)

func TestBFieldGet(t *testing.T) {
	cases := []struct {
		Name        string
		Input       []byte
		InputLength int
		InputIndex  int
		Want        bool
		WantError   bool
	}{
		{
			Name:        "Get bit from single byte",
			Input:       []byte{0b01000000},
			InputLength: 8,
			InputIndex:  1,
			Want:        true,
		},
		{
			Name:        "Error getting 8th bit of 7-bit byte",
			Input:       []byte{0},
			InputLength: 7, // Counting from 1, 2, ...
			InputIndex:  7, // Indexing from 0, 1, ...
			WantError:   true,
		},
		{
			Name:        "Get false bit from multi-byte array",
			Input:       []byte{0b00001010, 0b00000001},
			InputLength: 15,
			InputIndex:  11,
			Want:        false,
		},
		{
			Name:        "Error getting 14th bit from 11 bit field",
			Input:       []byte{0b00001010, 0b00010001},
			InputLength: 11,
			InputIndex:  14,
			WantError:   true,
		},
		{ // In case of manual creation of BField instead of NewBField, don't allow out of bounds read
			Name:        "Error getting beyond final byte when length too large",
			Input:       []byte{0, 0},
			InputLength: 17, // first bit of non-existent third byte, from 1
			InputIndex:  16, // first bit of non-existent byte, from 0
			WantError:   true,
		},
		{
			Name:        "Error getting bit from 0-length field",
			Input:       []byte{0},
			InputLength: 0,
			InputIndex:  0,
			WantError:   true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			bf := &BField{bs: c.Input, length: c.InputLength}
			got, err := bf.Get(c.InputIndex)
			switch {
			case c.WantError && err == nil:
				t.Fatal("Expected error, got none")
			case c.WantError:
				return
			case err != nil:
				t.Fatalf("Unexpected error: %s", err)
			case c.Want != got:
				t.Fatalf("Want %v, got %v", c.Want, got)
			default:
			}
		})
	}
}

func TestBFieldSet(t *testing.T) {
	cases := []struct {
		Name        string
		Input       []byte
		InputLength int
		InputIndex  int
		InputBool   bool
		Want        []byte
		WantError   bool
	}{
		{
			Name:        "Set bit in single byte",
			Input:       []byte{0b00001010},
			InputLength: 8,
			InputIndex:  1,
			InputBool:   true,
			Want:        []byte{0b01001010}, // change high bit 1 (from 0)
		},
		{
			Name:        "Error setting 8th bit of 7-bit byte",
			Input:       []byte{0},
			InputLength: 7, // Counting from 1, 2, ...
			InputIndex:  7, // Indexing from 0, 1, ...
			WantError:   true,
		},
		{
			Name:        "Set bit to false in multi-byte array",
			Input:       []byte{0b00001010, 0b00010001},
			InputLength: 15,
			InputIndex:  11,
			InputBool:   false,
			Want:        []byte{0b00001010, 0b00000001},
		},
		{
			Name:        "Error setting 14th bit in 11 bit field",
			Input:       []byte{0b00001010, 0b00010001},
			InputLength: 11,
			InputIndex:  14,
			WantError:   true,
		},
		{ // In case of manual creation of BField instead of NewBField, don't allow out of bounds read
			Name:        "Error setting beyond final byte when length too large",
			Input:       []byte{0, 0},
			InputLength: 17, // first bit of non-existent third byte, from 1
			InputIndex:  16, // first bit of non-existent byte, from 0
			WantError:   true,
		},
		{
			Name:        "Error setting bit in 0-length field",
			Input:       []byte{0},
			InputLength: 0,
			InputIndex:  0,
			WantError:   true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			bf := &BField{bs: c.Input, length: c.InputLength}
			err := bf.Set(c.InputIndex, c.InputBool)
			switch {
			case c.WantError && err == nil:
				t.Fatal("Expected error, got none")
			case c.WantError:
				return
			case err != nil:
				t.Fatalf("Unexpected error: %s", err)
			case !bytes.Equal(c.Want, bf.bs):
				t.Fatalf("Want %b, got %b", c.Want, bf.bs)
			default:
			}
		})
	}
}

func TestBFieldSwap(t *testing.T) {
	cases := []struct {
		Name        string
		Input       []byte
		InputLength int
		InputIndex  int
		Want        []byte
		WantError   bool
	}{
		{
			Name:        "Swap bit to true in single byte",
			Input:       []byte{0b00001010},
			InputLength: 8,
			InputIndex:  1,
			Want:        []byte{0b01001010}, // change high bit 1 (from 0)
		},
		{
			Name:        "Error setting 8th bit of 7-bit byte",
			Input:       []byte{0},
			InputLength: 7, // Counting from 1, 2, ...
			InputIndex:  7, // Indexing from 0, 1, ...
			WantError:   true,
		},
		{
			Name:        "Swap bit to false in multi-byte array",
			Input:       []byte{0b00001010, 0b00010001},
			InputLength: 15,
			InputIndex:  11,
			Want:        []byte{0b00001010, 0b00000001},
		},
		{
			Name:        "Error setting 14th bit in 11 bit field",
			Input:       []byte{0b00001010, 0b00010001},
			InputLength: 11,
			InputIndex:  14,
			WantError:   true,
		},
		{ // In case of manual creation of BField instead of NewBField, don't allow out of bounds read
			Name:        "Error setting beyond final byte when length too large",
			Input:       []byte{0, 0},
			InputLength: 17, // first bit of non-existent third byte, from 1
			InputIndex:  16, // first bit of non-existent byte, from 0
			WantError:   true,
		},
		{
			Name:        "Error setting bit in 0-length field",
			Input:       []byte{0},
			InputLength: 0,
			InputIndex:  0,
			WantError:   true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			bf := &BField{bs: c.Input, length: c.InputLength}
			err := bf.Swap(c.InputIndex)
			switch {
			case c.WantError && err == nil:
				t.Fatal("Expected error, got none")
			case c.WantError:
				return
			case err != nil:
				t.Fatalf("Unexpected error: %s", err)
			case !bytes.Equal(c.Want, bf.bs):
				t.Fatalf("Want %b, got %b", c.Want, bf.bs)
			default:
			}
		})
	}
	t.Run("Swapping all bits same as XOR", func(t *testing.T) {
		t.Parallel()
		bf := BField{
			bs:     []byte{0xAA, 0xBB, 0xCC},
			length: 24,
		}
		want := []byte{^byte(0xAA), ^byte(0xBB), ^byte(0xCC)}
		var err error
		for i := 0; i < 24; i++ {
			err = bf.Swap(i)
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
		}
		if !bytes.Equal(want, bf.bs) {
			t.Fatalf("Want %b, got %b", want, bf.bs)
		}
	})
}

func TestNewBitfield(t *testing.T) {
	cases := []struct {
		Name        string
		Input       []byte
		InputLength int
		Want        []byte
		WantError   bool
	}{
		{
			Name:        "Creating empty bitfield errors",
			Input:       []byte{},
			InputLength: 0,
			WantError:   true,
		},
		{
			Name:        "Creating non-empty aligned bitfield succeeds",
			Input:       []byte{'0', '1', '2', '3'},
			InputLength: 32,
			Want:        []byte{'0', '1', '2', '3'},
			WantError:   false,
		},
		{
			Name:        "Creating non-aligned bitfield succeeds",
			Input:       []byte{'0', '1', '2', '3'},
			InputLength: 31,
			Want:        []byte{'0', '1', '2', '3'},
			WantError:   false,
		},
		{
			Name:        "Creating bitfield with extra length errors",
			Input:       []byte{'0', '1', '2', '3'},
			InputLength: 33,
			WantError:   true,
		},
		{
			Name:        "Creating bitfield with extra byte(s) errors",
			Input:       []byte{'0', '1', '2', '3'},
			InputLength: 23, // last bit of '2'
			WantError:   true,
		},
		{
			Name:        "Creating bitfield with negative length errors",
			Input:       []byte{},
			InputLength: -1,
			WantError:   true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			bf, err := NewBitfield(c.Input, c.InputLength)
			switch {
			case c.WantError && err == nil:
				t.Fatal("Expected error, got none")
			case c.WantError:
				return
			case err != nil:
				t.Fatalf("Unexpected error: %s", err)
			case !bytes.Equal(c.Want, bf.bs): // Might as well check, but errors are what matters here
				t.Fatalf("Want %b, got %b", c.Want, bf.bs)
			default:
			}
		})
	}
}

func TestNewEmptyBitfield(t *testing.T) {
	cases := []struct {
		Name           string
		InputLength    int
		WantByteLength int
		WantError      bool
	}{
		{
			Name:           "Aligned bit length succeeds",
			InputLength:    64,
			WantByteLength: 8,
			WantError:      false,
		},
		{
			Name:           "Non-aligned bit length succeeds with correct byte length",
			InputLength:    65,
			WantByteLength: 9,
			WantError:      false,
		},
		{
			Name:        "Negative integer length errors",
			InputLength: -1,
			WantError:   true,
		},
		{
			Name:        "Zero length errors",
			InputLength: 0,
			WantError:   true,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			bf, err := NewEmptyBitfield(c.InputLength)
			switch {
			case c.WantError && err == nil:
				t.Fatal("Expected error, got none")
			case c.WantError:
				return
			case err != nil:
				t.Fatalf("Unexpected error: %s", err)
			case c.WantByteLength != len(bf.bs):
				t.Fatalf("Want byte length %d, got %d", c.WantByteLength, len(bf.bs))
			default:
			}
		})
	}
}
