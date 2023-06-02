package main

import (
	"reflect"
	"testing"
)

func TestParseTrackerResponse(t *testing.T) {
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
	// No Interval
	testInput := []byte(`d5:peersld4:peer8:testpeer2:ip7:1.2.3.44:porti3333eeee`)
	_, err := ParseTrackerResponse(testInput)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}
