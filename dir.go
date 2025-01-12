package vpk

import (
	"fmt"
	"os"
	"regexp"
)

var (
	reDirPath = regexp.MustCompile(`_(dir|\d{3}).vpk$`)
)

// Open a VPK file with multiple index files.
func OpenDir(path string) (*VPK, error) {
	find := reDirPath.FindString(path)
	if find == "" {
		return nil, ErrInvalidPath
	}

	noext := path[:len(path)-len(find)]

	dir := fmt.Sprintf("%s_dir.vpk", noext)

	v, err := OpenSingle(dir)
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, ErrInvalidVPKVersion
	}

	for i := 0; i <= 999; i++ {
		idx := fmt.Sprintf("%s_%03d.vpk", noext, i)

		ifs, err := os.Open(idx)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			defer v.Close()
			return nil, err
		}

		v.indexes = append(v.indexes, ifs)
	}

	return v, nil
}
