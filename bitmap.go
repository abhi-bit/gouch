package gouch

import (
	"bytes"
)

const BITMAP_SIZE int = (1024 / 8)

type Bitmap struct {
	data []byte
}

func CreateBitmap() *Bitmap {
	return &Bitmap{data: make([]byte, BITMAP_SIZE)}
}

func (b *Bitmap) SetBit(k uint64) {
	b.data[k/8] |= (1 << (k % 8))
}

func (b *Bitmap) ClearBit(k uint64) {
	b.data[k/8] &= ^(1 << (k % 8))
}

func (b *Bitmap) GetBit(k uint64) bool {
	return (b.data[k/8] & (1 << (k % 8))) != 0
}

func (b *Bitmap) Dump() []byte {
	return b.data
}

func UnionBitmap(b1 *Bitmap, b2 *Bitmap) {
	for i := 0; i < BITMAP_SIZE; i++ {
		b2.data[i] |= b1.data[i]
	}
}

func IntersectBitmap(b1 *Bitmap, b2 *Bitmap) {
	for i := 0; i < BITMAP_SIZE; i++ {
		b2.data[i] &= b1.data[i]
	}
}

func IsEqualBitmap(b1 *Bitmap, b2 *Bitmap) bool {
	return (bytes.Compare(b1.data, b2.data) == 0)
}
