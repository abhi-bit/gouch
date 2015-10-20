package gouch

import (
//"hash/crc32"
)

// ChunkLengthSize 32 bits long
const ChunkLengthSize int64 = 4

//ChunkCRCSize 32 bits long
const ChunkCRCSize int64 = 4

// attempt to read a chunk at the specified location
func (g *Gouch) readChunkAt(pos int64, header bool) ([]byte, int64, error) {
	var size uint32
	var n int64
	var err error
	// chunk starts with 8 bytes (32bit length, 32bit crc)
	if chunkPrefix, ok := eightByte.Get().([]byte); ok {

		n, err = g.readAt(chunkPrefix, 8, pos)
		if err != nil {
			return nil, int64(size), err
		}
		if n < ChunkLengthSize+ChunkCRCSize {
			return nil, int64(size), nil
		}

		size = decodeRaw31(chunkPrefix[0:ChunkLengthSize])
		//crc := decodeRaw32(chunkPrefix[ChunkLengthSize : ChunkLengthSize+ChunkCRCSize])
		eightByte.Put(chunkPrefix)
	}
	// size should at least be the size of the length field + 1 (for headers)
	if header && size < uint32(ChunkLengthSize+1) {
		return nil, int64(size), nil
	}
	if header {
		size -= uint32(ChunkLengthSize) // headers include the length of the hash, data does not
	}

	var data []byte
	var ok bool
	if data, ok = snappyDecodeChunk.Get().([]byte); ok {
		// skip the actual number of bytes read for the header (may be more than
		// header size if we crossed a block boundary)
		pos += n
		n, err = g.readAt(data, int64(size), pos)
		if uint32(n) < size {
			return nil, int64(size), nil
		}
	}

	// validate crc
	/*actualCRC := crc32.ChecksumIEEE(data)
	if actualCRC != crc {
		return nil, nil
	}*/

	return data, int64(size), nil
}

func (g *Gouch) readCompressedDataChunkAt(pos int64) ([]byte, error) {
	chunk, size, err := g.readChunkAt(pos, false)
	if err != nil {
		return nil, err
	}
	if len(chunk) == SnappyDecodeBufLen {
		snappyDecodeChunk.Put(chunk)
	}

	var decompressedChunk []byte
	var ok bool

	if decompressedChunk, ok = snappyDecodeChunk.Get().([]byte); ok {
		decompressedChunk, err = SnappyDecode(nil, chunk[:size])
		if err != nil {
			return nil, err
		}
	}
	return decompressedChunk, nil
}
