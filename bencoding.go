package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

func main() {
}

var ErrorEmpty = func() error { return errors.New("received empty input") }

var patInt = regexp.MustCompile(`^(?:(0)[^0-9]|(-?[1-9]\d*))`)

// ParseInt parses a *literal* integer value. It does NOT parse a Bencoded integer with i<int>e prefixing.
//
// "", -0, 00, 01, etc all produce errors.
func ParseInt(bs []byte) (any, []byte, error) {
	if len(bs) == 0 {
		return nil, bs, ErrorEmpty()
	}
	matches := patInt.FindSubmatch(bs)
	if matches == nil {
		return nil, bs, errors.New("ParseInt: no match found")
	}
	if len(matches) != 3 {
		return nil, bs, fmt.Errorf("expected exactly 3 matches, got %d", len(matches))
	}
	if len(matches[1]) != 0 {
		return 0, bs[1:], nil
	}
	data := matches[2]
	n, err := strconv.Atoi(string(data))
	if err != nil {
		return nil, bs, err
	}
	return n, bs[len(data):], nil
}

// Parse iINTe
func ParseInteger(bs []byte) (any, []byte, error) {
	rest, err := delim('i', bs)
	if err != nil {
		return nil, bs, err
	}
	i, rest, err := ParseInt(rest)
	if err != nil {
		return nil, bs, err
	}
	rest, err = delim('e', rest)
	if err != nil {
		return nil, bs, err
	}
	return i, rest, nil
}

// ParseLength parses a nonnegative integer (can be zero)
func ParseLength(bs []byte) (any, []byte, error) {
	if len(bs) == 0 {
		return nil, bs, ErrorEmpty()
	}
	rest := bs
	i, rest, err := ParseInt(rest)
	if err != nil {
		return nil, bs, err
	}
	// Check if negative
	n := i.(int)
	if n < 0 {
		return nil, bs, fmt.Errorf("expected nonnegative integer, got %d", n)
	}
	return i, rest, nil
}

func ParseString(bs []byte) (any, []byte, error) {
	// Parse length
	l, rest, err := ParseLength(bs)
	if err != nil {
		return nil, bs, err
	}
	length := l.(int)
	// Parse colon
	rest, err = delim(':', rest)
	if err != nil {
		return nil, bs, err
	}
	// Read length bytes
	if len(rest) < length {
		return nil, bs, fmt.Errorf("expected to read %d bytes, found %d", length, len(rest))
	}
	return string(rest[:length]), rest[length:], nil
}

func ParseList(bs []byte) (any, []byte, error) {
	// Parse l
	rest, err := delim('l', bs)
	if err != nil {
		return nil, bs, err
	}
	// Parse e (end) or value
	results := []any{}
	for len(rest) > 0 {
		if rest[0] == 'e' {
			// Create BList, trim e from rest, return.
			return results, rest[1:], nil
		}
		// Parse a term
		var next any // Prevent := below to avoid shadowing rest
		next, rest, err = Parse(rest)
		if err != nil {
			return nil, bs, err
		}
		results = append(results, next)
	}
	return nil, bs, errors.New("received incomplete list")
}

func ParseDict(bs []byte) (any, []byte, error) {
	rest, err := delim('d', bs)
	if err != nil {
		return nil, bs, err
	}
	results := make(map[string]any)
	for len(rest) > 0 {
		if rest[0] == 'e' { // End of dict
			// Create BMap, trim e from rest, return.
			return results, rest[1:], nil
		}
		// Parse a key string
		var keyString any // Don't use := in order to avoid shadowing rest below
		keyString, rest, err = ParseString(rest)
		if err != nil {
			return nil, bs, fmt.Errorf("failed to parse key: %s", err)
		}
		key, ok := keyString.(string)
		if !ok {
			return nil, bs, fmt.Errorf("expected string key, got %T", keyString)
		}
		// Parse a value
		var value any // Don't use := in order to avoid shadowing rest below
		value, rest, err = Parse(rest)
		if err != nil {
			return nil, bs, fmt.Errorf("failed to parse value for key %s: %s", key, err)
		}
		//TODO what if key already exists? What does spec say?
		results[key] = value
	}
	return nil, bs, errors.New("reached EOF without completing dictionary")
}

func Parse(bs []byte) (any, []byte, error) {
	if len(bs) == 0 {
		return nil, bs, ErrorEmpty()
	}
	switch bs[0] {
	case 'i': // integer
		return ParseInteger(bs)
	case 'l': // list
		return ParseList(bs)
	case 'd': // dict
		return ParseDict(bs)
	default: // string or error
		if '0' <= bs[0] && bs[0] <= '9' {
			return ParseString(bs)
		}
		return nil, bs, fmt.Errorf("expected start of term, got %x", bs[0])
	}
}

// delim tries to parse a single byte b from bs.
//
// It always returns rest even on error.  It doesn't return the parsed value, since
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
