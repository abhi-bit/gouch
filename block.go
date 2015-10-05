package gouch

import (
	"fmt"
)

const (
	BLOCK_SIZE        int64 = 4096
	BLOCK_MARKER_SIZE int64 = 1
	BLOCK_DATA        byte  = 0
	BLOCK_HEADER      byte  = 1
	BLOCK_INVALID     byte  = 0xff
)

func (g *Gouch) seekPreviousBlockFrom(pos int64) (int64, byte, error) {
	//pos -= 1
	pos -= pos % BLOCK_SIZE
	for ; pos >= 0; pos -= BLOCK_SIZE {
		fmt.Println("Seeking previous block!")
		var err error
		buf := make([]byte, 1)
		n, err := g.ops.ReadAt(g.file, buf, pos)
		if n != 1 || err != nil {
			return -1, BLOCK_INVALID, err
		}
		if buf[0] == BLOCK_HEADER {
			return pos, BLOCK_HEADER, nil
		} else if buf[0] == BLOCK_DATA {
			return pos, BLOCK_DATA, nil
		} else {
			return -1, BLOCK_INVALID, nil
		}
	}
	return -1, BLOCK_INVALID, nil
}

func (g *Gouch) seekLastHeaderBlockFrom(pos int64) (int64, error) {
	var blockType byte
	var err error
	for pos, blockType, err = g.seekPreviousBlockFrom(pos); blockType != BLOCK_HEADER; pos, blockType, err = g.seekPreviousBlockFrom(pos) {
		fmt.Println("Inside seekLastHeaderBLockFrom function call")
		if err != nil {
			return -1, err
		}
	}
	return pos, nil
}

func (g *Gouch) readAt(buf []byte, pos int64) (int64, error) {
	bytesReadSoFar := int64(0)
	bytesSkipped := int64(0)
	numBytesToRead := int64(len(buf))
	readOffset := pos
	for numBytesToRead > 0 {
		//var err error
		bytesTillNextBlock := BLOCK_SIZE - (readOffset % BLOCK_SIZE)
		if bytesTillNextBlock == BLOCK_SIZE {
			readOffset++
			bytesTillNextBlock--
			bytesSkipped++
		}
		bytesToReadThisPass := bytesTillNextBlock
		if bytesToReadThisPass > numBytesToRead {
			bytesToReadThisPass = numBytesToRead
		}
		n, err := g.ops.ReadAt(g.file, buf[bytesReadSoFar:bytesReadSoFar+bytesToReadThisPass], readOffset)
		if err != nil {
			fmt.Errorf("Failed to read from offset: %d\n", readOffset)
			return -1, err
		}
		readOffset += int64(n)
		bytesReadSoFar += int64(n)
		numBytesToRead -= int64(n)
		if int64(n) < bytesToReadThisPass {
			fmt.Printf("Bytes read: %d\n", bytesToReadThisPass)
			return bytesReadSoFar, nil
		}
	}
	return bytesReadSoFar + bytesSkipped, nil
}
