package gouch

import (
//"hash/crc32"
)

// ChunkLengthSize 32 bits long
const ChunkLengthSize int64 = 4

//ChunkCRCSize 32 bits long
const ChunkCRCSize int64 = 4

// attempt to read a chunk at the specified location
func (g *Gouch) readChunkAt(pos int64, header bool) ([]byte, error) {
	// chunk starts with 8 bytes (32bit length, 32bit crc)
	chunkPrefix := make([]byte, ChunkLengthSize+ChunkCRCSize)
	n, err := g.readAt(chunkPrefix, pos)
	if err != nil {
		return nil, err
	}
	if n < ChunkLengthSize+ChunkCRCSize {
		return nil, nil
	}

	size := decodeRaw31(chunkPrefix[0:ChunkLengthSize])
	//crc := decodeRaw32(chunkPrefix[ChunkLengthSize : ChunkLengthSize+ChunkCRCSize])

	// size should at least be the size of the length field + 1 (for headers)
	if header && size < uint32(ChunkLengthSize+1) {
		return nil, nil
	}
	if header {
		size -= uint32(ChunkLengthSize) // headers include the length of the hash, data does not
	}

	data := make([]byte, size)
	pos += n // skip the actual number of bytes read for the header (may be more than header size if we crossed a block boundary)
	n, err = g.readAt(data, pos)
	if uint32(n) < size {
		return nil, nil
	}

	// validate crc
	/*actualCRC := crc32.ChecksumIEEE(data)
	if actualCRC != crc {
		return nil, nil
	}*/

	return data, nil
}

func (g *Gouch) readCompressedDataChunkAt(pos int64) ([]byte, error) {
	chunk, err := g.readChunkAt(pos, false)
	if err != nil {
		return nil, err
	}

	decompressedChunk, err := SnappyDecode(nil, chunk)
	if err != nil {
		return nil, err
	}
	return decompressedChunk, nil
}
