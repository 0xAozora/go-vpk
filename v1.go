package vpk

import (
	"bufio"
	"encoding/binary"
	"io"
)

func openVPK_v1(fs FileReader, buffer []byte) (*VPK, error) {
	if _, err := fs.Read(buffer[:4]); err != nil {
		return nil, err
	}

	v := &VPK{
		stream:     fs,
		version:    1,
		headerSize: 4 * 3,

		TreeSize: int32(binary.LittleEndian.Uint32(buffer[:4])),

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
