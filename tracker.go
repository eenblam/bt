package main

import "errors"

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

func ParseTrackerResponse(bs []byte) (*TrackerResponse, error) {
	tr, err := FromBencode[TrackerResponse](bs)
	if err != nil {
		return nil, err
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
	//TODO parse peers into addresses
	return &tr, nil
}
