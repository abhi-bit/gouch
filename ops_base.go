package gouch

import (
	"os"

	"github.com/golang/snappy"
)

//BaseOps struct for doing file operations
type BaseOps struct{}

//NewBaseOps returns new BaseOps struct ptr
func NewBaseOps() *BaseOps {
	return &BaseOps{}
}

//OpenFile opens an index file with defined perms
func (b *BaseOps) OpenFile(name string, flag int, perm os.FileMode) (file *os.File, err error) {
	return os.OpenFile(name, flag, perm)
}

//ReadAt reads file at different marker
func (b *BaseOps) ReadAt(f *os.File, bytes []byte, off int64) (n int, err error) {
	return f.ReadAt(bytes, off)
}

//WriteAt writes at specific marker
func (b *BaseOps) WriteAt(f *os.File, bytes []byte, off int64) (n int, err error) {
	return f.WriteAt(bytes, off)
}

//GotoEOF traverses file from end
func (b *BaseOps) GotoEOF(f *os.File) (n int64, err error) {
	return f.Seek(0, os.SEEK_END)
}

//Sync syncs file to disk
func (b *BaseOps) Sync(f *os.File) error {
	return f.Sync()
}

//Close closes index file
func (b *BaseOps) Close(f *os.File) error {
	return f.Close()
}

//SnappyEncode encodes source byte array using snappy
func SnappyEncode(dst, src []byte) []byte {
	return snappy.Encode(dst, src)
}

//SnappyDecode decodes source byte array using snappy
func SnappyDecode(dst, src []byte) ([]byte, error) {
	return snappy.Decode(dst, src)
}

//SnappyDecodeLength dumps the len of snappy encoded data
func SnappyDecodeLength(bytes []byte) (int, error) {
	return snappy.DecodedLen(bytes)
}
