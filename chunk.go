package gouch

import (
	"fmt"
	"hash/crc32"
)

const (
	CHUNK_LENGTH_SIZE int64 = 4
	CHUNK_CRC_SIZE    int64 = 4
)

func (g *Gouch) readChunkAt(pos int64, header bool) ([]byte, error) {
	chunkPrefix := make([]byte, CHUNK_LENGTH_SIZE+CHUNK_CRC_SIZE)
	n, err := g.readAt(chunkPrefix, pos)

	if err != nil {
		fmt.Errorf("Failed in readAt function call, err: %v\n", err)
		return nil, err
	}

	size := decode_raw31(chunkPrefix[0:CHUNK_LENGTH_SIZE])
	crc := decode_raw32(chunkPrefix[CHUNK_LENGTH_SIZE : CHUNK_LENGTH_SIZE+CHUNK_CRC_SIZE])
	fmt.Printf("Chunk size: %d\n", size)

	if header && size < uint32(CHUNK_LENGTH_SIZE+1) {
		fmt.Println("Chunk size too small")
		return nil, nil
	}

	if header {
		size -= uint32(CHUNK_LENGTH_SIZE)
	}

	data := make([]byte, size)
	pos += n
	n, err = g.readAt(data, pos)
	if uint32(n) < size {
		fmt.Println("Chunk data less than size")
		return nil, nil
	}

	actualCRC := crc32.ChecksumIEEE(data)
	if actualCRC != crc {
		fmt.Println("Invalid chunk bad crc32")
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
