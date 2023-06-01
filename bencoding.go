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

type BencodingType int

const (
	Integer BencodingType = iota
	List
	String
	Dictionary
)

// type VList []Value
// type VMap map[string]Value
// type Value [string | int | VList | VMap]
type Value interface {
	Int() int
	List() []Value
	Map() map[string]Value
	//String() string
	String() []byte
	Type() BencodingType
}

type value struct {
	i int
	l []Value
	m map[string]Value
	//s string
	// bs is a "byte string"
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

// func BString(bs []byte) Value {
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

type Result struct {
	//Value Value
	//Value interface{}
	Value Value
	Rest  []byte
	// Error *Error ???
	Error error
}

/*
type Parser func(input []byte) Result

func Success(payload interface{}, remaining []byte) Result {
	return Result{Value: payload, Rest: remaining}
}

func Fail(err error, input []byte) Result {
	return Result{Error: err, Rest: input}
}

var patLength = regexp.MustCompile(`^(\d+):`)
*/

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
		return Result{Rest: bs, Error: errors.New("no match found")}
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

func delim(b byte, bs []byte) ([]byte, error) {
	if len(bs) == 0 {
		return bs, ErrorEmpty()
	}
	if bs[0] != b {
		return bs, fmt.Errorf("want %b, got %b", b, bs[0])
	}
	return bs[1:], nil
}
