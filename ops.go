package gouch

import (
	"os"
)

type Ops interface {
	OpenFile(name string, flag int, perm os.FileMode) (file *os.File, err error)
	ReadAt(f *os.File, b []byte, off int64) (n int, err error)
	WriteAt(f *os.File, b []byte, off int64) (n int, err error)
	GotoEOF(f *os.File) (n int64, err error)
	Sync(f *os.File) error
	//SnappyEncode(dst, src []byte) []byte
	//SnappyDecode(dst, src []byte) ([]byte, error)
	Close(f *os.File) error
}
