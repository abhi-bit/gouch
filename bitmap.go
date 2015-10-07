package gouch

import (
	"bytes"
)

//BitmapSize length of byte array to store bitmap
const BitmapSize int = (1024 / 8)

//Bitmap data struct
type Bitmap struct {
	data []byte
}

//CreateBitmap creates bitmap
func CreateBitmap() *Bitmap {
	return &Bitmap{data: make([]byte, BitmapSize)}
}

//SetBit sets a specici bit in bitmap
func (b *Bitmap) SetBit(k uint64) {
	b.data[k/8] |= (1 << (k % 8))
}

//ClearBit clears a specific bit in bitmap
func (b *Bitmap) ClearBit(k uint64) {
	b.data[k/8] &= ^(1 << (k % 8))
}

//GetBit gets a specific bit in bitmap
func (b *Bitmap) GetBit(k uint64) bool {
	return (b.data[k/8] & (1 << (k % 8))) != 0
}

//Dump dumps the bitmap data struct
func (b *Bitmap) Dump() []byte {
	return b.data
}

//UnionBitmap ORs two input bitmaps
func UnionBitmap(b1 *Bitmap, b2 *Bitmap) {
	for i := 0; i < BitmapSize; i++ {
		b2.data[i] |= b1.data[i]
	}
}

//IntersectBitmap ANDs two input bitmap
func IntersectBitmap(b1 *Bitmap, b2 *Bitmap) {
	for i := 0; i < BitmapSize; i++ {
		b2.data[i] &= b1.data[i]
	}
}

//IsEqualBitmap checks if two input bitmaps are the same
func IsEqualBitmap(b1 *Bitmap, b2 *Bitmap) bool {
	return (bytes.Compare(b1.data, b2.data) == 0)
}
