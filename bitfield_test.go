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

func TestNextFalse(t *testing.T) {
	cases := []struct {
		Name        string
		InputBytes  []byte
		InputLength int
		// What to initialize to. -1 to count from beginning.
		SetNextFalse int
		WantIndex    int
		WantDone     bool
		WantError    bool
	}{
		{
			Name:         "Succeeds on zero byte array",
			InputBytes:   []byte{0, 0, 0},
			InputLength:  24,
			WantIndex:    0,
			WantDone:     false,
			SetNextFalse: -1,
		},
		{
			Name:         "Succeeds on non-initial byte",
			InputBytes:   []byte{255, 0b10101111, 255},
			InputLength:  24,
			WantIndex:    9,
			WantDone:     false,
			SetNextFalse: -1,
		},
		{
			Name:         "Done when done, single byte",
			InputBytes:   []byte{255},
			InputLength:  8,
			WantIndex:    8,
			WantDone:     true,
			SetNextFalse: 8, // we've already counted this
		},
		{
			Name:         "Done when done",
			InputBytes:   []byte{255, 255},
			InputLength:  16,
			WantIndex:    16,
			WantDone:     true,
			SetNextFalse: 16, // we've already counted this
		},
		{
			Name:         "Doesn't overread into spare bits",
			InputBytes:   []byte{255, 0b11110000},
			InputLength:  12,
			WantIndex:    12,
			WantDone:     true,
			SetNextFalse: -1, // fine to start from the start
		},
		{
			// This one is sadly implementation-specific but I think important for a specific bug
			Name:         "Resuming from mid-byte doesn't lead to skipped bits in later byte",
			InputBytes:   []byte{255, 0b01110000},
			InputLength:  16,
			WantIndex:    8,
			WantDone:     false,
			SetNextFalse: 5, // start at 5th bit of first byte, find bit in 2nd bit of second byte
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			bf, err := NewBitfield(c.InputBytes, c.InputLength)
			bf.nextFalse = c.SetNextFalse
			if err != nil {
				t.Fatalf("Unexpected error initializing BField: %s", err)
			}
			gotNext, gotDone := bf.NextFalse()
			switch {
			case c.WantError && err == nil:
				t.Fatal("Expected error, got none")
			case c.WantError:
				return
			case err != nil:
				t.Fatalf("Unexpected error: %s", err)
			case c.WantIndex != gotNext:
				t.Fatalf("Want index %d, got %d", c.WantIndex, gotNext)
			case c.WantDone != gotDone:
				t.Fatalf("Want done=%v, got done=%v", c.WantDone, gotDone)
			default:
			}
		})
	}

	t.Run("Reset nextFalse correctly for Set", func(t *testing.T) {
		t.Parallel()
		bf, err := NewBitfield([]byte{255, 0b11110111}, 16)
		if err != nil {
			t.Fatalf("Unexpected error initializing BField: %s", err)
		}
		// Initial check; ensure nextFalse=12
		nf, done := bf.NextFalse()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 12 {
			t.Fatalf("Expected nextFalse to be 12, got %d", nf)
		}
		// Set an earlier 1 to a 0
		if err = bf.Set(5, false); err != nil {
			t.Fatalf("Unexpected error in first Set: %s", err)
		}
		nf, done = bf.NextFalse()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextFalse to be 5 after first Set, got %d", nf)
		}
		// Set a *later* 1 to a 0
		if err = bf.Set(6, false); err != nil {
			t.Fatalf("Unexpected error in second Set: %s", err)
		}
		nf, done = bf.NextFalse()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextFalse to still be 5 after second swap, got %d", nf)
		}
	})

	t.Run("Reset nextFalse correctly for Swap", func(t *testing.T) {
		t.Parallel()
		bf, err := NewBitfield([]byte{255, 0b11110111}, 16)
		if err != nil {
			t.Fatalf("Unexpected error initializing BField: %s", err)
		}
		// Initial check; ensure nextFalse=12
		nf, done := bf.NextFalse()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 12 {
			t.Fatalf("Expected nextFalse to be 12, got %d", nf)
		}
		// Swap an earlier 1 to a 0
		if err = bf.Swap(5); err != nil {
			t.Fatalf("Unexpected error in first swap: %s", err)
		}
		nf, done = bf.NextFalse()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextFalse to be 5 after first Swap, got %d", nf)
		}
		// Swap a *later* 1 to a 0
		if err = bf.Swap(6); err != nil {
			t.Fatalf("Unexpected error in second Swap: %s", err)
		}
		nf, done = bf.NextFalse()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextFalse to still be 5 after second Swap, got %d", nf)
		}
	})
}

func TestNextTrue(t *testing.T) {
	cases := []struct {
		Name        string
		InputBytes  []byte
		InputLength int
		// What to initialize to. -1 to count from beginning.
		SetNextTrue int
		WantIndex   int
		WantDone    bool
		WantError   bool
	}{
		{
			Name:        "Succeeds on all-ones byte array",
			InputBytes:  []byte{255, 255, 255},
			InputLength: 24,
			WantIndex:   0,
			WantDone:    false,
			SetNextTrue: -1,
		},
		{
			Name:        "Succeeds on non-initial byte",
			InputBytes:  []byte{0, 0b01010000, 0},
			InputLength: 24,
			WantIndex:   9,
			WantDone:    false,
			SetNextTrue: -1,
		},
		{
			Name:        "Done when done, single byte",
			InputBytes:  []byte{0},
			InputLength: 8,
			WantIndex:   8,
			WantDone:    true,
			SetNextTrue: 8, // we've already counted this
		},
		{
			Name:        "Done when done",
			InputBytes:  []byte{0, 0},
			InputLength: 16,
			WantIndex:   16,
			WantDone:    true,
			SetNextTrue: 16, // we've already counted this
		},
		{
			Name:        "Doesn't overread into spare bits",
			InputBytes:  []byte{0, 0b00000000},
			InputLength: 12,
			WantIndex:   12, // as opposed to reading past length to end
			WantDone:    true,
			SetNextTrue: -1, // fine to start from the start
		},
		{
			// This one is sadly implementation-specific but I think important for a specific bug
			Name:        "Resuming from mid-byte doesn't lead to skipped bits in later byte",
			InputBytes:  []byte{0, 0b10001111},
			InputLength: 16,
			WantIndex:   8,
			WantDone:    false,
			SetNextTrue: 5, // start at 5th bit of first byte, find bit in 2nd bit of second byte
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			bf, err := NewBitfield(c.InputBytes, c.InputLength)
			bf.nextTrue = c.SetNextTrue
			if err != nil {
				t.Fatalf("Unexpected error initializing BField: %s", err)
			}
			gotNext, gotDone := bf.NextTrue()
			switch {
			case c.WantError && err == nil:
				t.Fatal("Expected error, got none")
			case c.WantError:
				return
			case err != nil:
				t.Fatalf("Unexpected error: %s", err)
			case c.WantIndex != gotNext:
				t.Fatalf("Want index %d, got %d", c.WantIndex, gotNext)
			case c.WantDone != gotDone:
				t.Fatalf("Want done=%v, got done=%v", c.WantDone, gotDone)
			default:
			}
		})
	}

	t.Run("Reset nextTrue correctly for Set", func(t *testing.T) {
		t.Parallel()
		bf, err := NewBitfield([]byte{0, 0b00001000}, 16)
		if err != nil {
			t.Fatalf("Unexpected error initializing BField: %s", err)
		}
		// Initial check; ensure nextTrue=12
		nf, done := bf.NextTrue()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 12 {
			t.Fatalf("Expected nextTrue to be 12, got %d", nf)
		}
		// Set an earlier 1 to a 0
		if err = bf.Set(5, true); err != nil {
			t.Fatalf("Unexpected error in first Set: %s", err)
		}
		nf, done = bf.NextTrue()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextTrue to be 5 after first Set, got %d", nf)
		}
		// Set a *later* 1 to a 0
		if err = bf.Set(6, true); err != nil {
			t.Fatalf("Unexpected error in second Set: %s", err)
		}
		nf, done = bf.NextTrue()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextTrue to still be 5 after second swap, got %d", nf)
		}
	})

	t.Run("Reset nextTrue correctly for Swap", func(t *testing.T) {
		t.Parallel()
		bf, err := NewBitfield([]byte{0, 0b00001000}, 16)
		if err != nil {
			t.Fatalf("Unexpected error initializing BField: %s", err)
		}
		// Initial check; ensure nextTrue=12
		nf, done := bf.NextTrue()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 12 {
			t.Fatalf("Expected nextTrue to be 12, got %d", nf)
		}
		// Swap an earlier 1 to a 0
		if err = bf.Swap(5); err != nil {
			t.Fatalf("Unexpected error in first swap: %s", err)
		}
		nf, done = bf.NextTrue()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextTrue to be 5 after first Swap, got %d", nf)
		}
		// Swap a *later* 1 to a 0
		if err = bf.Swap(6); err != nil {
			t.Fatalf("Unexpected error in second Swap: %s", err)
		}
		nf, done = bf.NextTrue()
		if done {
			t.Fatal("Unexpected done=true")
		}
		if nf != 5 {
			t.Fatalf("Expected nextTrue to still be 5 after second Swap, got %d", nf)
		}
	})
}
