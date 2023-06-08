package bt

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// See https://www.bittorrent.org/beps/bep_0003.html

type MetaInfo struct {
	Announce string `json:"announce"`
	Info     Info   `json:"info"`
	// Could just do InfoSha1
	InfoShaSum [sha1.Size]byte `json:"-"`
}

type Info struct {
	// name: maps to a UTF-8 encoded string which is the suggested name to save the file (or directory) as. It is purely advisory.
	// In the single file case, the name key is the name of a file, in the muliple file case, it's the name of a directory.
	Name string `json:"name,omitempty"`
	// piece length: the number of bytes in each piece the file is split into. (Last may be truncated.)
	PieceLength int `json:"piece length"`
	// pieces: string whose length is a multiple of 20, subdivided into strings of length 20, each a SHA1 hash of the piece at the corresponding index.
	Pieces       [][]byte `json:"-"`
	PiecesString string   `json:"pieces"`
	// Length OR Files. Check if Files is nil?
	Length *int       `json:"length,omitempty"`
	Files  []FileInfo `json:"files,omitempty"`
}

type FileInfo struct {
	// The length of the file, in bytes.
	Length int `json:"length"`
	// If length zero, error
	// A list of UTF-8 encoded strings corresponding to subdirectory names, the last of which is the actual file name (a zero length list is an error case).
	Path []string `json:"path"`
}

func LoadMetaInfoFromFile(filename string) (*MetaInfo, error) {
	bs, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseMetaInfo(bs)
}

func ParseMetaInfo(bs []byte) (*MetaInfo, error) {
	m, err := parseMetaInfo(bs)
	if err != nil {
		return nil, err
	}
	// length OR files
	if m.Info.Length != nil && m.Info.Files != nil {
		return nil, errors.New("MetaInfo:Info: info dict must have exactly one of \"length\" or \"files\", found both")
	}
	if m.Info.Length == nil && m.Info.Files == nil {
		return nil, errors.New("MetaInfo:Info: info dict must have exactly one of \"length\" or \"files\", found neither")
	}
	// if Files, confirm each Path []string is nonempty
	if m.Info.Files != nil {
		for _, file := range m.Info.Files {
			if len(file.Path) == 0 {
				return nil, errors.New("MetaInfo:Info:Files: each file must have a nonempty path")
			}
		}
	}
	// Parse PiecesString (a string of length n*20) to Pieces
	pieces := []byte(m.Info.PiecesString)
	if len(pieces)%20 != 0 {
		return nil, fmt.Errorf("MetaInfo:Info:Pieces: length of string \"pieces\" must be a multiple of 20, got length=%d", len(pieces))
	}
	nPieces := len(pieces) / 20
	m.Info.Pieces = make([][]byte, nPieces)
	for i := 0; i < nPieces; i++ {
		m.Info.Pieces[i] = pieces[i*20 : (i+1)*20]
	}
	return m, nil
}

// Just here for debugging at the moment
func (m *MetaInfo) String() string {
	pieces := make([]string, len(m.Info.Pieces))
	for i, p := range m.Info.Pieces {
		pieces[i] = fmt.Sprintf("%x", p)
	}

	length := ""
	if m.Info.Length != nil {
		length = fmt.Sprint(*m.Info.Length)
	}
	fmt.Println(length)

	files := ""
	if m.Info.Files != nil {
		for _, f := range m.Info.Files {
			files += fmt.Sprintf("\n\t(%d bytes) %s", f.Length, strings.Join(f.Path, "/"))
		}
	} else {
		files = "none"
	}

	return strings.Join([]string{
		fmt.Sprintf("MetaInfo.Announce: %s", m.Announce),
		fmt.Sprintf("MetaInfo.InfoSha1Sum (hex): %x", m.InfoShaSum),
		fmt.Sprintf("MetaInfo.Info.Name: %s", m.Info.Name),
		fmt.Sprintf("MetaInfo.Info.Piece length: %d", m.Info.PieceLength),
		fmt.Sprintf("MetaInfo.Info.Pieces (length): %d", len(m.Info.Pieces)),
		//fmt.Sprintf("MetaInfo.Info.Pieces (hex):\n\t%s", strings.Join(pieces, "\n\t")),
		"MetaInfo.Info.Pieces (hex): (omitted)",
		fmt.Sprintf("MetaInfo.Info.Length: %s", length),
		fmt.Sprintf("MetaInfo.Info.Files: %s", files),
	}, "\n\t")
}

// parseMetaInfo handles the lower-level parsing of a metainfo file,
// in which we need to be sure to extract the raw bencoded version of the info dict
// in order to pass it to the tracker.
func parseMetaInfo(bs []byte) (*MetaInfo, error) {
	// "a" for announce should place its key first, before "info" key
	if len(bs) == 0 {
		return nil, ErrorEmpty()
	}
	// strip d from beginning
	if bs[0] != 'd' {
		return nil, fmt.Errorf("MetaInfo: expected metainfo to begin with 'd', got %b", bs[0])
	}
	bs = bs[1:]
	// read 8:announce
	got, rest, err := ParseString(bs)
	if err != nil {
		return nil, err
	}
	if got != "announce" {
		return nil, fmt.Errorf("MetaInfo: expected \"announce\", got %s", got)
	}
	// read some string (value of announce)
	announceAny, rest, err := ParseString(rest)
	if err != nil {
		return nil, err
	}
	announce, ok := announceAny.(string)
	if !ok {
		return nil, fmt.Errorf("MetaInfo: expected announce value to have type string, got %T", announceAny)
	}

	for _, k := range []string{"comment", "created by", "creation date"} {
		got, rest, err = ParseString(rest)
		if err != nil {
			return nil, err
		}
		if got != k {
			return nil, fmt.Errorf("MetaInfo: expected '%s', got '%s'", k, got)
		}
		_, rest, err = Parse(rest)
		if err != nil {
			return nil, err
		}
	}

	// read 4:info key (as string)
	got, rest, err = ParseString(rest)
	if err != nil {
		return nil, err
	}
	if got != "info" {
		return nil, fmt.Errorf("MetaInfo: expected \"info\", got %s", got)
	}

	// d8:announce6:value4:infodINFOee
	// 8:announce6:value4:infodINFOee
	// 6:value4:infodINFOee
	// 4:infodINFOee
	// dINFOee
	// dINFOe <- Need to extract this to get SHA1
	// Now left with dINFOee. Strip last e.
	if len(rest) == 0 {
		return nil, errors.New("MetaInfo: unexpected EOF")
	}
	if rest[len(rest)-1] != 'e' {
		return nil, fmt.Errorf("MetaInfo: expected 'e' at end of metainfo, got %d", rest[len(rest)-1])
	}
	rest = rest[:len(rest)-1]
	rawInfo := rest
	// Parse the rest as dict, coerce into an &Info{} (value of info)
	parsedInfo, rest, err := ParseDict(rest)
	if err != nil {
		return nil, err
	}
	// We should now be done
	if len(rest) != 0 {
		return nil, fmt.Errorf("expected EOF after parsing Info dict, but found %d bytes", len(rest))
	}

	// Extract an actual struct
	js, err := json.Marshal(parsedInfo)
	if err != nil {
		return nil, err
	}
	var info Info
	err = json.Unmarshal(js, &info)
	if err != nil {
		return nil, err
	}

	// store raw bytes into MetaInfo struct as well
	return &MetaInfo{
		Announce:   announce,
		Info:       info,
		InfoShaSum: sha1.Sum(rawInfo),
	}, nil
}
