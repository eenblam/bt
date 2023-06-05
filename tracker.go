package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

/*
Tracker responses are bencoded dictionaries.

If key `failure reason` is present, no other keys are needed.
`failure reason` "maps to a human readable string which explains why the query failed."

Otherwise, it must have two keys: interval and peers.
`interval`` maps to the number of seconds the downloader should wait between regular rerequests.
`peers` maps to a list of dictionaries corresponding to peers,
each of which contains the keys peer id, ip, and port.
`peer` maps to the peer's self-selected ID.
`ip` maps to IP address or dns name as a string.
`port` maps to port number.

Note that downloaders may rerequest on nonscheduled times if an event happens or they need more peers.

More commonly is that trackers return a compact representation of the peer list, see BEP 23.
https://www.bittorrent.org/beps/bep_0023.html
*/

type TrackerResponse struct {
	// Don't love defining these as pointer,
	// but I'm not sure how best to check if they were provided otherwise.
	Reason   *string `json:"failure reason,omitempty"`
	Interval *int    `json:"interval,omitempty"`
	Peers    []Peer  `json:"peers,omitempty"`
}

type Peer struct {
	Peer string `json:"peer"` // string???
	//TODO add a non-JSON address generated from these?
	// i.e. ParseTrackerResponse should try to create it, and error if the address can't be created.
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// Parses a TrackerResponse from the JSON representation of a bencoded dictionary.
func ParseClassicTrackerResponse(jsonBytes []byte) (*TrackerResponse, error) {
	var tr TrackerResponse
	err := json.Unmarshal(jsonBytes, &tr)
	if err != nil {
		return &tr, err
	}
	// Interval and Peers must both be non-nil if Reason is nil
	if tr.Reason == nil {
		if tr.Interval == nil {
			return nil, errors.New("TrackerResponse: Interval cannot be nil when Reason is nil")
		}
		if tr.Peers == nil {
			return nil, errors.New("TrackerResponse: Peers cannot be nil when Reason is nil")
		}
	}
	//TODO parse peers into net.IP addresses
	return &tr, nil
}

type CompactTrackerResponse struct {
	Reason   *string `json:"failure reason,omitempty"`
	Interval *int    `json:"interval,omitempty"`
	Peers    string  `json:"peers,omitempty"`
}

// Parses a compact TrackerResponse from the JSON representation of a bencoded dictionary.
func ParseCompactTrackerResponse(jsonBytes []byte) (*TrackerResponse, error) {
	var tr CompactTrackerResponse
	err := json.Unmarshal(jsonBytes, &tr)
	if err != nil {
		return nil, err
	}
	if tr.Reason != nil {
		// Don't care about unpacking Peers if failure
		return &TrackerResponse{Reason: tr.Reason}, nil
	}
	// Interval and Peers must both be non-nil if Reason is nil
	if tr.Reason == nil {
		if tr.Interval == nil {
			return nil, errors.New("TrackerResponse: Interval cannot be nil when Reason is nil")
		}
		if tr.Peers == "" {
			return nil, errors.New("TrackerResponse: Peers cannot be missing when Reason is nil")
		}
	}

	// Parse Peers string into a list
	if len(tr.Peers)%6 != 0 {
		return nil, fmt.Errorf("expected Peers string to be divisible by 6, got %d", len(tr.Peers))
	}
	// Parse
	peers := []Peer{}
	var (
		p          Peer
		ipBytes    = make([]byte, 4)
		ipUint32   uint32
		ip         string
		portUint16 uint16
		port       int
	)
	for i := 0; i < len(tr.Peers); i += 6 {
		// Need to read as network byte order (big endian) and convert to little endian
		ipUint32 = binary.BigEndian.Uint32([]byte(tr.Peers[i : i+4]))
		binary.LittleEndian.PutUint32(ipBytes, ipUint32)
		ip = net.IPv4(ipBytes[0], ipBytes[1], ipBytes[2], ipBytes[3]).String()

		portUint16 = binary.BigEndian.Uint16([]byte(tr.Peers[i+4 : i+6]))
		port = int(portUint16)
		//TODO convert from bytes to address

		p = Peer{
			Peer: "", // Ignored in compact format
			IP:   ip,
			Port: port,
		}
		peers = append(peers, p)
	}

	return &TrackerResponse{Interval: tr.Interval, Peers: peers}, nil
}

type TrackerResponsePartial struct {
	// Don't love defining these as pointer,
	// but I'm not sure how best to check if they were provided otherwise.
	Reason   *string `json:"failure reason,omitempty"`
	Interval *int    `json:"interval,omitempty"`
	Peers    *string `json:"peers,omitempty"`
}

// Parse either:
// Parse TrackerResponsePartial
// Can I check the first byte of a RawMessage? String -> Compact, List -> Classic

func ParseTrackerResponse(bs []byte) (*TrackerResponse, error) {
	v, _, err := Parse(bs)
	if err != nil {
		return nil, err
	}
	js, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Try Classic
	tr, err := ParseClassicTrackerResponse(js)
	if err == nil {
		return tr, nil
	}
	// Failed - try compact
	return ParseCompactTrackerResponse(js)
}
