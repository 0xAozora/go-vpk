package vpk

import (
	"bufio"
	"encoding/binary"
	"io"
)

func openVPK_v2(fs FileReader, buffer []byte) (*VPK, error) {
	if _, err := fs.Read(buffer[:4*5]); err != nil {
		return nil, err
	}

	v := &VPK{
		stream:     fs,
		version:    2,
		headerSize: 4 * 7,

		TreeSize:              int32(binary.LittleEndian.Uint32(buffer[:4])),
		FileDataSectionSize:   int32(binary.LittleEndian.Uint32(buffer[4:8])),
		ArchiveMD5SectionSize: int32(binary.LittleEndian.Uint32(buffer[8:12])),
		OtherMD5SectionSize:   int32(binary.LittleEndian.Uint32(buffer[12:16])),
		SignatureSectionSize:  int32(binary.LittleEndian.Uint32(buffer[16:20])),

		PathMap: make(map[string]*Entry),
	}

	reader := bufio.NewReader(io.LimitReader(fs, int64(v.TreeSize)))

	if err := treeReader(v, reader, buffer, v.addFile); err != nil {
		defer v.Close()
		return nil, err
	}

	// We should have read exactly .treeSize bytes and therefore hit EOF
	if _, err := reader.ReadByte(); err != io.EOF {
		defer v.Close()
		return nil, ErrWrongHeaderSize
	}

	return v, nil
}
