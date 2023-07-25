package bt

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
		return out, fmt.Errorf("failed to read random bytes: %w", err)
	}
	return out, nil
}

type Downloader struct {
	MetaInfo  MetaInfo
	PeerId    [20]byte
	LocalPort int
	// Where pieces will be downloaded to
	PiecesDir   string
	isMultifile bool
	downloaded  int
	uploaded    int
	// The number of bytes this peer still has to download, encoded in base ten ascii.
	// Note that this can't be computed from downloaded and the file length since it might be a resume,
	// and there's a chance that some of the downloaded data failed an integrity check and had to be re-downloaded.
	left     int
	listener *net.TCPListener
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

	piecesDir, err := SetupStorage(fmt.Sprintf("%x", m.InfoShaSum))
	if err != nil {
		return nil, err
	}

	return &Downloader{
		MetaInfo:    *m,
		PeerId:      peerId,
		PiecesDir:   piecesDir,
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
	v.Set("port", fmt.Sprint(d.LocalPort))
	v.Set("uploaded", fmt.Sprint(d.uploaded))
	v.Set("downloaded", fmt.Sprint(d.downloaded))
	v.Set("left", fmt.Sprint(d.left))
	v.Set("event", "started")
	return v.Encode(), nil
}

func (d *Downloader) QueryTracker() (*TrackerResponse, error) {
	q, err := d.MakeTrackerQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to create tracker query: %w", err)
	}
	u := fmt.Sprintf("%s?%s", d.MetaInfo.Announce, q)
	r, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", u, err)
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: expected 200 OK, got %s", u, r.Status)
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("GET %s: reading body: %w", u, err)
	}
	return ParseTrackerResponse(data)
}

// If needed, creates directories required for download based on environment variable.
// No error if directories already exist.
func SetupStorage(downloadName string) (string, error) {
	//TODO consider different env variable name and default name. Maybe use a flag instead?
	workRoot := os.Getenv("BT_WORKROOT")
	if workRoot == "" {
		workRoot = "./bt-work/"
	}
	downloadDir := filepath.Join(workRoot, "download", downloadName)
	log.Printf("Using download directory %s", downloadDir)
	err := os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", err
	} else if os.IsExist(err) {
		log.Printf("Download directory already available: %s", downloadDir)
	}
	// Directory for building output file from pieces
	err = os.MkdirAll(filepath.Join(workRoot, "build"), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return "", err
	}
	return downloadDir, nil
}

// Listen cycles through ports 6881 to 6889, erroring if it can't bind any
//
// See BEP0003
func (d *Downloader) Listen() error {
	// Cycle through
	var listener *net.TCPListener // so we don't have to shadow err below
	var err error
	for port := 6881; port < 6890; port++ {
		// Try to listen on port
		listener, err = net.ListenTCP("tcp", &net.TCPAddr{
			IP:   nil,
			Port: port,
			Zone: "", //TODO support IPv6
		})
		if err == nil {
			log.Printf("listening on port %d", port)
			d.listener = listener
			d.LocalPort = port
			return nil
		}
	}
	return fmt.Errorf("couldn't listen on ports 6881-6889. Last error: %w", err)
}

// ListenPort attempts to listen specifically on a given port
func (d *Downloader) ListenPort(port int) error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   nil,
		Port: port,
		Zone: "", //TODO support IPv6
	})
	if err != nil {
		return fmt.Errorf("couldn't listen on port %d: %w", port, err)
	}
	log.Printf("listening on port %d", port)
	d.listener = listener
	d.LocalPort = port
	return nil
}

// Close underlying TCPListener
func (d *Downloader) Close() {
	d.listener.Close()
}
