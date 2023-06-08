package bt

import (
	"crypto/rand"
	"fmt"
	"net/url"
)

// Using Azureus-style peer id.
// `EE` for eenblam. BT and EB are already in use.
// Remaining 4 chars are version.
// See https://wiki.theory.org/BitTorrentSpecification#peer_id
var PeerPrefix = [8]byte{'-', 'E', 'E', '0', '0', '0', '0', '-'}

// Generate a 20-byte peerId
func GenPeerId() ([20]byte, error) {
	//TODO should be unique on this system! How to ensure no collision?
	out := [20]byte{}
	copy(out[:8], PeerPrefix[:]) // ignore n, we have length guarantees here
	rest := out[8:]
	_, err := rand.Read(rest)
	if err != nil { // only check err, since n = len(rest) iff err == nil
		return out, fmt.Errorf("failed to read random bytes: %s", err)
	}
	return out, nil
}

type Downloader struct {
	MetaInfo    *MetaInfo
	PeerId      [20]byte
	ListenPort  int
	isMultifile bool
	downloaded  int
	uploaded    int
	// The number of bytes this peer still has to download, encoded in base ten ascii.
	// Note that this can't be computed from downloaded and the file length since it might be a resume,
	// and there's a chance that some of the downloaded data failed an integrity check and had to be re-downloaded.
	left int
}

func NewDownloader(filename string) (*Downloader, error) {
	m, err := LoadMetaInfoFromFile(filename)
	if err != nil {
		return nil, err
	}
	peerId, err := GenPeerId()
	if err != nil {
		return nil, err
	}
	//TODO get a port to listen on
	port := 9999
	return &Downloader{
		MetaInfo:    m,
		PeerId:      peerId,
		ListenPort:  port,
		isMultifile: m.Info.Files != nil,
	}, nil
}

func (d *Downloader) MakeTrackerQuery() (string, error) {
	// info_hash
	// peer_id - generate string of length 20 at random. Generate this elsewhere later.
	// ip - ignore?
	// port - need to bind one
	// uploaded - 0 for now
	// downloaded - 0 for now
	// left - whatever amounts to "all"
	// event - for first request, should be "started"
	v := url.Values{}
	v.Set("info_hash", string(d.MetaInfo.InfoShaSum[:]))
	v.Set("peer_id", string(d.PeerId[:]))
	//v.Set("ip", "")
	v.Set("port", fmt.Sprint(d.ListenPort))
	v.Set("uploaded", fmt.Sprint(d.uploaded))
	v.Set("downloaded", fmt.Sprint(d.downloaded))
	v.Set("left", fmt.Sprint(d.left))
	v.Set("event", "started")
	return v.Encode(), nil
}
