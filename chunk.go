package gouch

import (
	"hash/crc32"
)

const (
	CHUNK_LENGTH_SIZE int64 = 4
	CHUNK_CRC_SIZE    int64 = 4
)

// attempt to read a chunk at the specified location
func (g *Gouch) readChunkAt(pos int64, header bool) ([]byte, error) {
	// chunk starts with 8 bytes (32bit length, 32bit crc)
	chunkPrefix := make([]byte, CHUNK_LENGTH_SIZE+CHUNK_CRC_SIZE)
	n, err := g.readAt(chunkPrefix, pos)
	if err != nil {
		return nil, err
	}
	if n < CHUNK_LENGTH_SIZE+CHUNK_CRC_SIZE {
		return nil, nil
	}

	size := decode_raw31(chunkPrefix[0:CHUNK_LENGTH_SIZE])
	crc := decode_raw32(chunkPrefix[CHUNK_LENGTH_SIZE : CHUNK_LENGTH_SIZE+CHUNK_CRC_SIZE])

	// size should at least be the size of the length field + 1 (for headers)
	if header && size < uint32(CHUNK_LENGTH_SIZE+1) {
		return nil, nil
	}
	if header {
		size -= uint32(CHUNK_LENGTH_SIZE) // headers include the length of the hash, data does not
	}

	data := make([]byte, size)
	pos += n // skip the actual number of bytes read for the header (may be more than header size if we crossed a block boundary)
	n, err = g.readAt(data, pos)
	if uint32(n) < size {
		return nil, nil
	}

	// validate crc
	actualCRC := crc32.ChecksumIEEE(data)
	if actualCRC != crc {
		return nil, nil
	}

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
