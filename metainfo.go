package main

import (
	"errors"
	"os"
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
	Pieces string `json:"pieces"`
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
	return &m, nil
}
