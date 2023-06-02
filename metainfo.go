package main

import (
	"errors"
	"fmt"
)

type MetaInfo struct {
	Announce string
	Info     Info
}

type Info struct {
	// The name key maps to a UTF-8 encoded string which is the suggested name to save the file (or directory) as. It is purely advisory.
	// In the single file case, the name key is the name of a file, in the muliple file case, it's the name of a directory.
	Name        string
	PieceLength int // `piece length`
	// String whose length is a multiple of 20, subdivided into strings of length 20, each a SHA1 hash of the piece at the corresponding index.
	Pieces string
	// Length OR Files. Check if Files is nil?
	Length int
	Files  []FileInfo
}

type FileInfo struct {
	// The length of the file, in bytes.
	Length int
	// If length zero, error
	// A list of UTF-8 encoded strings corresponding to subdirectory names, the last of which is the actual file name (a zero length list is an error case).
	Path []string
}

func ReadMetainfo(f string) Value {
	// Read from file
	// Try to parse
	return &value{}
}

func GetAnnounce(v Value) (string, error) {
	m := v.Map()
	if m == nil {
		return "", fmt.Errorf("expected Dict, got %s", v.Type())
	}
	announce, ok := m["announce"]
	if !ok {
		return "", errors.New("expected 'announce'")
	}
	if announce.Type() != String {
		return "", fmt.Errorf("expected String, got %s", announce.Type())
	}
	return string(announce.String()), nil
}

func GetInfo(v Value) (*Info, error) {
	m := v.Map()
	if m == nil {
		return nil, fmt.Errorf("expected Dict, got %s", v.Type())
	}

	// Name
	nameV, ok := m["nameV"]
	if !ok {
		return nil, errors.New("expected 'nameV'")
	}
	if nameV.Type() != String {
		return nil, fmt.Errorf("expected String, got %s", nameV.Type())
	}
	name := nameV.String()
	// PieceLength
	pieceLengthV, ok := m["piece length"]
	if !ok {
		return nil, errors.New("expected 'piece length'")
	}
	if pieceLengthV.Type() != Integer {
		return nil, fmt.Errorf("expected Integer, got %s", pieceLengthV.Type())
	}
	pieceLength := pieceLengthV.Int()
	if pieceLength < 0 {
		return nil, fmt.Errorf("want nonnegative piece length, got %d", pieceLength)
	}
	// Pieces
	piecesV, ok := m["piecesV"]
	if !ok {
		return nil, errors.New("expected 'piecesV'")
	}
	if piecesV.Type() != String {
		return nil, fmt.Errorf("expected String, got %s", piecesV.Type())
	}
	pieces := piecesV.String()
	if len(pieces)%20 != 0 {
		return nil, fmt.Errorf("expected length of pieces to be multple of 20, got %d", len(pieces))
	}
	// Length OR Files
	lengthV, lengthOk := m["length"]
	filesV, filesOk := m["files"]
	if lengthOk && filesOk {
		return nil, errors.New("got both 'length' and 'files' keys")
	}
	if !lengthOk && !filesOk {
		return nil, errors.New("expected 'length' or 'files, got neither")
	}
	var length int
	var files []FileInfo
	// Length
	if lengthOk {
		if lengthV.Type() != Integer {
			return nil, fmt.Errorf("expected length to be Integer, got %d", lengthV.Type())
		}
		length = lengthV.Int()
	}
	// Files
	if filesOk {
		if filesV.Type() != Dictionary {
			return nil, fmt.Errorf("expected length to be Dictionary, got %d", filesV.Type())
		}
		var err error // don't use :=, would shadow files
		files, err = parseFilesValue(filesV)
		if err != nil {
			return nil, err
		}
	}
	return &Info{
		Name:        string(name),
		PieceLength: pieceLength,
		Pieces:      string(pieces),
		Length:      length,
		Files:       files,
	}, nil
}

func parseFilesValue(v Value) ([]FileInfo, error) {
	if v.Type() != List {
		return nil, fmt.Errorf("expected List, got %d", v.Type())
	}
	l := v.List()
	out := []FileInfo{}
	for _, val := range l {
		if val.Type() != Dictionary {
			return nil, fmt.Errorf("expected list entry to be Dict, got %d", val.Type())
		}
		d := val.Map()
		// length
		lengthV, ok := d["length"]
		if !ok {
			return nil, errors.New("expected 'length' key in file info, got none")
		}
		if lengthV.Type() != Integer {
			return nil, fmt.Errorf("expected list entry to be Dict, got %d", lengthV.Type())
		}
		length := lengthV.Int()
		if length < 0 {
			return nil, fmt.Errorf("got negative length %d", length)
		}
		// path
		pathV, ok := d["path"]
		if !ok {
			return nil, errors.New("expected 'path' key in file info, got none")
		}
		if pathV.Type() != List {
			return nil, fmt.Errorf("expected 'path' value to be List, got %d", pathV.Type())
		}
		pathsV := pathV.List()
		paths := []string{}
		for _, p := range pathsV {
			if p.Type() != String {
				return nil, fmt.Errorf("expected path list entry to be String, got %d", p.Type())
			}
			paths = append(paths, string(p.String()))

		}
		out = append(out, FileInfo{Length: length, Path: paths})
	}
	return out, nil
}
