package vpk

import "io"

type FileReader interface {
	io.Reader
	io.ReaderAt
	io.Closer
}
