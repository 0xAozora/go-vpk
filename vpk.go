package vpk

import (
	"io"
	"os"
)

type ReaderAtCloser interface {
	io.ReaderAt
	io.Closer
}

type VPK struct {
	stream     ReaderAtCloser
	indexes    []ReaderAtCloser
	version    int
	headerSize int

	// The size, in bytes, of the directory tree
	TreeSize int32

	// How many bytes of file content are stored in this VPK file (0 in CSGO)
	FileDataSectionSize int32

	// The size, in bytes, of the section containing MD5 checksums for external archive content
	ArchiveMD5SectionSize int32

	// The size, in bytes, of the section containing MD5 checksums for content in this file (should always be 48)
	OtherMD5SectionSize int32

	// The size, in bytes, of the section containing the public key and signature. This is either 0 (CSGO & The Ship) or 296 (HL2, HL2:DM, HL2:EP1, HL2:EP2, HL2:LC, TF2, DOD:S & CS:S)
	SignatureSectionSize int32

	Entries []*Entry
	PathMap map[string]*Entry
}

func (v *VPK) addFile(e *Entry) {
	v.Entries = append(v.Entries, e)
	v.PathMap[e.Filename()] = e
	e.parent = v
}

func (v *VPK) Open(path string) (io.Reader, error) {
	entry, ok := v.PathMap[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return entry.Open()
}

func (v *VPK) Close() error {
	v.stream.Close()

	for _, f := range v.indexes {
		f.Close()
	}

	return nil
}
