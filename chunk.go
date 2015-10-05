package gouch

import (
	"fmt"
)

const CHUNK_LENGTH_SIZE int64 = 4
const CHUNK_CRC_SIZE int64 = 4

func (g *Gouch) readChunkAt(pos int64, header bool) ([]byte, error) {
	chunkPrefix := make([]byte, CHUNK_LENGTH_SIZE+CHUNK_CRC_SIZE)
	n, err := g.readAt(chunkPrefix, pos)

	if err != nil {
		fmt.Errorf("Failed in readAt function call, err: %v\n", err)
		return nil, err
	}

	size := decode_raw31(chunkPrefix[0:CHUNK_LENGTH_SIZE])

	if header {
		size -= uint32(CHUNK_LENGTH_SIZE)
	}

	data := make([]byte, size)
	pos += n
	n, err = g.readAt(data, pos)
	return chunkPrefix, nil
}
