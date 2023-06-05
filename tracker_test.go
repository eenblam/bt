package main

import (
	"reflect"
	"testing"
)

func TestParseClassicTrackerResponse(t *testing.T) {
	testInput := []byte(`d8:intervali10e5:peersld4:peer8:testpeer2:ip7:1.2.3.44:porti3333eeee`)
	ten := 10
	testWant := &TrackerResponse{
		Interval: &ten,
		Peers: []Peer{
			{
				Peer: "testpeer",
				IP:   "1.2.3.4",
				Port: 3333,
			},
		},
	}
	got, err := ParseTrackerResponse(testInput)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !reflect.DeepEqual(testWant, got) {
		t.Fatalf("want %#v, got %#v", testWant, got)
	}
}

func TestParseTrackerResponseFailure(t *testing.T) {
	// No Interval. Neither sub-parser should pass.
	testInput := []byte(`d5:peersld4:peer8:testpeer2:ip7:1.2.3.44:porti3333eeee`)
	_, err := ParseTrackerResponse(testInput)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}

func TestParseCompactTrackerResponse(t *testing.T) {
	peersBytes := []byte{'1', '2', ':', // 12: (length)
		//0x01, 0x02, 0x03, 0x04, 0x0d, 0x05, // 1.2.3.4:3333
		//0x10, 0x20, 0x30, 0x40, 0x05, 0x0d, // 16.32.48.64:1293
		4, 3, 2, 1, 0x0d, 0x05, // 1.2.3.4:3333 in network byte order
		1, 2, 3, 4, 0x05, 0x0d, // 4.3.2.1:1293 in network byte order
	}
	testInput := append([]byte(`d8:intervali10e5:peers`)[:],
		peersBytes[:]...)
	testInput = append(testInput, []byte(`e`)...) // end d
	ten := 10
	testWant := &TrackerResponse{
		Interval: &ten,
		Peers: []Peer{
			{
				IP:   "1.2.3.4",
				Port: 3333,
			},
			{
				IP:   "4.3.2.1",
				Port: 1293,
			},
		},
	}
	got, err := ParseTrackerResponse(testInput)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !reflect.DeepEqual(testWant, got) {
		t.Fatalf("\nwant\n\t%#v\ngot\n\t%#v", testWant, got)
	}
}
