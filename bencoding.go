package main

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

func main() {
}

var ErrorEmpty = func() error { return errors.New("received empty input") }

type BencodingType int

const (
	Integer BencodingType = iota
	List
	String
	Dictionary
)

func (b BencodingType) String() string {
	switch b {
	case Integer:
		return "Integer"
	case List:
		return "List"
	case String:
		return "String"
	case Dictionary:
		return "Dictionary"
	default:
		return fmt.Sprintf("UNKNOWN_BENCODING_TYPE %d", b)
	}
}

type Value interface {
	Int() int
	List() []Value
	Map() map[string]Value
	//String() string
	String() []byte
	Type() BencodingType
	Equal(v Value) bool
}

type value struct {
	//TODO need to support arbitrary integer size
	i int
	l []Value
	m map[string]Value
	//s string
	// bs is a "byte string", per the spec
	bs []byte
	t  BencodingType
}

func BInt(n int) Value {
	return &value{i: n, t: Integer}
}

func BList(l []Value) Value {
	return &value{l: l, t: List}
}

func BMap(m map[string]Value) Value {
	return &value{m: m, t: Dictionary}
}

// []byte, not actual string, per spec
func BString(bs []byte) Value {
	return &value{bs: bs, t: String}
}

func (v *value) Int() int {
	return v.i
}

func (v *value) List() []Value {
	return v.l
}

func (v *value) Map() map[string]Value {
	return v.m
}

// func (v *value) String() string {
func (v *value) String() []byte {
	return v.bs
}

func (v *value) Type() BencodingType {
	return v.t
}

func (v *value) Equal(u Value) bool {
	if v.Type() != u.Type() {
		return false
	}
	switch v.Type() {
	case Integer:
		return v.Int() == u.Int()
	case String:
		return bytes.Equal(v.String(), u.String())
	case List:
		a, b := v.List(), u.List()
		if len(a) != len(b) {
			return false
		}
		for i := 0; i < len(a); i++ {
			x, y := a[i], b[i]
			if !x.Equal(y) {
				return false
			}
		}
		return true
	case Dictionary:
		a, b := v.Map(), u.Map()
		if len(a) != len(b) {
			return false
		}
		for k, vv := range a {
			uu, ok := b[k]
			if !ok {
				return false
			}
			if !vv.Equal(uu) {
				return false
			}
		}
		return true
	default:
	}
	panic(fmt.Sprintf("unexpected BencodingType %d", v.Type()))
}

type Result struct {
	Value Value
	Rest  []byte
	Error error
}

var patInt = regexp.MustCompile(`^(?:(0)[^0-9]|(-?[1-9]\d*))`)

// ParseInt parses an integer value to a Result.
// It does NOT parse a Bencoded integer with i<int>e prefixing.
//
// "", -0, 00, 01, etc all produce errors.
func ParseInt(bs []byte) Result {
	if len(bs) == 0 {
		return Result{Rest: bs, Error: ErrorEmpty()}
	}
	matches := patInt.FindSubmatch(bs)
	if matches == nil {
		return Result{Rest: bs, Error: errors.New("ParseInt: no match found")}
	}
	if len(matches) != 3 {
		fmt.Println(matches)
		return Result{Rest: bs, Error: fmt.Errorf("expected exactly 3 matches, got %d", len(matches))}
	}
	if len(matches[1]) != 0 {
		return Result{Value: BInt(0), Rest: bs[1:]}
	}
	data := matches[2]
	n, err := strconv.Atoi(string(data))
	if err != nil {
		return Result{Rest: bs, Error: err}
	}
	return Result{Value: BInt(n), Rest: bs[len(data):]}
}

// Parse iINTe
func ParseInteger(bs []byte) Result {
	rest, err := delim('i', bs)
	if err != nil {
		return Result{Rest: bs, Error: err}
	}
	r := ParseInt(rest)
	if r.Error != nil {
		return Result{Rest: bs, Error: r.Error}
	}
	rest, err = delim('e', r.Rest)
	if err != nil {
		return Result{Rest: bs, Error: err}
	}
	return Result{Value: r.Value, Rest: rest}
}

// ParseLength parses a nonnegative integer (can be zero)
func ParseLength(bs []byte) Result {
	if len(bs) == 0 {
		return Result{Rest: bs, Error: ErrorEmpty()}
	}
	rest := bs
	r := ParseInt(rest)
	if r.Error != nil {
		return r
	}
	// Check if negative
	n := r.Value.Int()
	if n < 0 {
		return Result{Rest: bs, Error: fmt.Errorf("expected nonnegative integer, got %d", n)}
	}
	return r
}

func ParseString(bs []byte) Result {
	// Parse length
	lr := ParseLength(bs)
	if lr.Error != nil {
		return lr
	}
	length := lr.Value.Int()
	// Parse colon
	rest, err := delim(':', lr.Rest)
	if err != nil {
		return Result{Rest: bs, Error: err}
	}
	// Read length bytes
	if len(rest) < length {
		return Result{Rest: bs, Error: fmt.Errorf("expected to read %d bytes, found %d", length, len(rest))}
	}
	return Result{Value: BString(rest[:length]), Rest: rest[length:]}
}

func ParseList(bs []byte) Result {
	// Parse l
	rest, err := delim('l', bs)
	if err != nil {
		return Result{Rest: bs, Error: err}
	}
	// Parse e (end) or value
	results := []Value{}
	for len(rest) > 0 {
		switch rest[0] {
		case 'e':
			// Create BList, trim e from rest, return.
			return Result{Value: BList(results), Rest: rest[1:]}
		default:
			// Parse a term
			next := Term(rest)
			if next.Error != nil {
				return Result{Rest: bs, Error: next.Error}
			}
			results = append(results, next.Value)
			rest = next.Rest
		}
	}
	return Result{Rest: bs, Error: errors.New("received incomplete list")}
}

func ParseDict(bs []byte) Result {
	rest, err := delim('d', bs)
	if err != nil {
		return Result{Rest: bs, Error: err}
	}
	results := make(map[string]Value)
	for len(rest) > 0 {
		switch rest[0] {
		case 'e': // End of dict
			// Create BMap, trim e from rest, return.
			return Result{Value: BMap(results), Rest: rest[1:]}
		default:
		}
		// Parse a key string
		keyResult := ParseString(rest)
		if keyResult.Error != nil {
			return Result{Rest: bs, Error: fmt.Errorf("failed to parse key: %s", keyResult.Error)}
		}
		//TODO should we instead use map[[]byte]Value? Is that hashable?
		key := string(keyResult.Value.String())
		// Parse a value
		valueResult := Term(keyResult.Rest)
		if valueResult.Error != nil {
			return Result{Rest: bs, Error: fmt.Errorf("failed to parse value for key %s: %s", key, valueResult.Error)}
		}
		//TODO what if key already exists? What does spec say?
		results[key] = valueResult.Value
		rest = valueResult.Rest
	}
	return Result{Rest: bs, Error: errors.New("reached EOF without completing dictionary")}
}

func Term(bs []byte) Result {
	if len(bs) == 0 {
		return Result{Rest: bs, Error: ErrorEmpty()}
	}
	switch bs[0] {
	case 'i': // integer
		return ParseInteger(bs)
	case 'l': // list
		return ParseList(bs)
	case 'd': // dict
		return ParseDict(bs)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // string (encountered length)
		return ParseString(bs)
	default: // error
		return Result{Rest: bs, Error: fmt.Errorf("expected start of term, got %x", bs[0])}
	}
}

// delim tries to parse a single byte b from bs.
//
// It always returns rest even on error.  It doesn't return a Result, since
// 1. the delimiter is assumed to be markup only
// 2. the caller will want to return a different rest on error more than 50% of the time
func delim(b byte, bs []byte) ([]byte, error) {
	if len(bs) == 0 {
		return bs, ErrorEmpty()
	}
	if bs[0] != b {
		return bs, fmt.Errorf("want %x, got %x", b, bs[0])
	}
	return bs[1:], nil
}
