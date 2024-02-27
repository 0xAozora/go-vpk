package vpk

import (
	"io"
	"regexp"
	"strings"
)

type Entry struct {
	parent *VPK

	ext  string
	path string
	file string

	// A 32bit CRC of the file's data.
	CRC uint32

	// The number of bytes contained in the index file.
	PreloadBytes uint16

	// A zero based index of the archive this file's data is contained in.
	// If 0x7fff, the data follows the directory.
	ArchiveIndex uint16

	// If ArchiveIndex is 0x7fff, the offset of the file data relative to the end of the directory (see the header for more details).
	// Otherwise, the offset of the data from the start of the specified archive.
	EntryOffset uint32

	// If zero, the entire file is stored in the preload data.
	// Otherwise, the number of bytes stored starting at EntryOffset.
	EntryLength uint32
}

func (e *Entry) Filename() string {
	var parts []string

	if e.path != " " {
		parts = append(parts, e.path)
	}

	if e.file != " " {
		if e.ext != " " {
			parts = append(parts, e.file+"."+e.ext)
		} else {
			parts = append(parts, e.file)
		}
	} else {
		if e.ext != " " {
			parts = append(parts, "."+e.ext)
		} else {
			parts = append(parts, "")
		}
	}

	return strings.Join(parts, "/")
}

func (e *Entry) Basename() string {
	if e.file != " " {
		if e.ext != " " {
			return e.file + "." + e.ext
		} else {
			return e.file
		}
	} else if e.ext != " " {
		return "." + e.ext
	}
	return ""
}

func (e *Entry) Path() string {
	if e.path == " " {
		return ""
	}
	return e.path
}

func (e *Entry) Length() uint32 {
	return e.EntryLength
}

var (
	reInvalidSegmentUnix = regexp.MustCompile(`^(\.\.?)$|\0`)
	reInvalidSegmentWin  = regexp.MustCompile(`^(?i)(\.\.?|CON|PRN|AUX|NUL|COM\d|LPT\d|)$|[\0-\x1f<>:"\\|?*]|[\s\.]$`)
	reRootWindows        = regexp.MustCompile(`^(?i)([A-Z]:|[\\/])`)
	reInvalidFilename    = regexp.MustCompile(`[/]`)
	reInvalidExt         = regexp.MustCompile(`[\./]`)
)

func (e *Entry) FilenameSafeUnix() bool {
	full := e.Filename()

	if full == "" || strings.HasPrefix(full, "/") || reInvalidFilename.MatchString(e.file) || reInvalidExt.MatchString(e.ext) {
		return false
	}

	for _, part := range strings.Split(full, "/") {
		if reInvalidSegmentUnix.MatchString(part) {
			return false
		}
	}

	return true
}

func (e *Entry) FilenameSafeWindows() bool {
	full := e.Filename()

	if full == "" || reRootWindows.MatchString(full) || reInvalidFilename.MatchString(e.file) || reInvalidExt.MatchString(e.ext) {
		return false
	}

	for _, part := range strings.Split(full, "/") {
		if reInvalidSegmentWin.MatchString(part) {
			return false
		}
	}

	return true
}

func (e *Entry) Open() (io.ReadCloser, error) {
	if e.ArchiveIndex == 0x7fff {
		return &EntryReader{
			fs:     e.parent.stream,
			offset: int64(e.parent.headerSize) + int64(e.parent.TreeSize) + int64(e.EntryOffset),
			size:   int64(e.EntryLength),
		}, nil
	}

	if e.ArchiveIndex >= uint16(len(e.parent.indexes)) {
		return nil, ErrInvalidArchiveIndex
	}

	return &EntryReader{
		fs:     e.parent.indexes[e.ArchiveIndex],
		offset: int64(e.EntryOffset),
		size:   int64(e.EntryLength),
	}, nil
}
