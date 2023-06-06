package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
)

// See https://www.bittorrent.org/beps/bep_0003.html

type MetaInfo struct {
	Announce string `json:"announce"`
	Info     Info   `json:"info"`
}

type Info struct {
	// The name key maps to a UTF-8 encoded string which is the suggested name to save the file (or directory) as. It is purely advisory.
	// In the single file case, the name key is the name of a file, in the muliple file case, it's the name of a directory.
	Name        string `json:"name,omitempty"`
	PieceLength int    `json:"piece length"`
	// String whose length is a multiple of 20, subdivided into strings of length 20, each a SHA1 hash of the piece at the corresponding index.
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
	m, err := FromBencode[MetaInfo](bs)
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
	return &m, nil
}

// Just here for debugging at the moment
func (m *MetaInfo) String() string {
	pieces := make([]string, len(m.Info.Pieces))
	for i, p := range m.Info.Pieces {
		pieces[i] = base64.StdEncoding.EncodeToString(p)
	}
	return strings.Join([]string{
		fmt.Sprintf("Announce: %s", m.Announce),
		fmt.Sprintf("Name: %s", m.Info.Name),
		fmt.Sprintf("Piece length: %d", m.Info.PieceLength),
		//fmt.Sprintf("Pieces:\n\t%s", strings.Join(pieces, "\n\t")),
		fmt.Sprintf("Length: %d", *m.Info.Length),
		//m.Info.Files
	}, "\n")
}
