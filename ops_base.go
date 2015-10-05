package gouch

import (
	"fmt"
	"os"

	"github.com/golang/snappy"
)

type BaseOps struct{}

func NewBaseOps() *BaseOps {
	return &BaseOps{}
}

func (b *BaseOps) OpenFile(name string, flag int, perm os.FileMode) (file *os.File, err error) {
	return os.OpenFile(name, flag, perm)
}

func (b *BaseOps) ReadAt(f *os.File, bytes []byte, off int64) (n int, err error) {
	n, err = f.ReadAt(bytes, off)
	fmt.Printf("ReadAt n: %v, err: %v\n", n, err)
	return n, err
}

func (b *BaseOps) WriteAt(f *os.File, bytes []byte, off int64) (n int, err error) {
	return f.WriteAt(bytes, off)
}

func (b *BaseOps) GotoEOF(f *os.File) (n int64, err error) {
	return f.Seek(0, os.SEEK_END)
}

func (b *BaseOps) Sync(f *os.File) error {
	return f.Sync()
}

func (b *BaseOps) Close(f *os.File) error {
	return f.Close()
}

func SnappyEncode(dst, src []byte) []byte {
	return snappy.Encode(dst, src)
}

func SnappyDecode(dst, src []byte) ([]byte, error) {
	return snappy.Decode(dst, src)
}

func SnappyDecodeLength(bytes []byte) (int, error) {
	return snappy.DecodedLen(bytes)
}
