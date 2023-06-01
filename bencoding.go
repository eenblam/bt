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
	String() string
	Type() BencodingType
}

type value struct {
	i int
	l []Value
	m map[string]Value
	s string
	t BencodingType
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

func BString(s string) Value {
	return &value{s: s, t: String}
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

func (v *value) String() string {
	return v.s
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
	rest := bs
	if len(rest) == 0 {
		return Result{Rest: bs, Error: ErrorEmpty()}
	}
	if rest[0] != 'i' {
		return Result{Rest: bs, Error: fmt.Errorf("expected i, got %s", string(rest[0]))}
	}
	rest = rest[1:]
	r := ParseInt(rest)
	if r.Error != nil {
		return Result{Rest: bs, Error: r.Error}
	}
	rest = r.Rest
	if rest[0] != 'e' {
		return Result{Rest: bs, Error: fmt.Errorf("expected e, got %s", string(rest[0]))}
	}
	rest = rest[1:]
	return Result{Value: r.Value, Rest: rest}
}

// ParseLength parses a nonnegative integer (can be zero)
func ParseLength(bs []byte) Result {
	return Result{}
}
