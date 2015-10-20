package gouch

//BlockSize BTree block size
const BlockSize int64 = 4096

//BlockData Data block
const BlockData byte = 0

//BlockHeader Header block
const BlockHeader byte = 1

//BlockInvalid marker for invalid btree block
const BlockInvalid byte = 0xff

func (g *Gouch) seekPreviousBlockFrom(pos int64) (int64, byte, error) {
	pos--
	pos -= pos % BlockSize
	for ; pos >= 0; pos -= BlockSize {
		var err error
		buf := make([]byte, 1)
		n, err := g.ops.ReadAt(g.file, buf, pos)
		if n != 1 || err != nil {
			return -1, BlockInvalid, err
		}
		if buf[0] == BlockHeader {
			return pos, BlockHeader, nil
		} else if buf[0] == BlockData {
			return pos, BlockData, nil
		} else {
			return -1, BlockInvalid, nil
		}
	}
	return -1, BlockInvalid, nil
}

func (g *Gouch) seekLastHeaderBlockFrom(pos int64) (int64, error) {
	var blockType byte
	var err error
	for pos, blockType, err = g.seekPreviousBlockFrom(pos); blockType != BlockHeader; pos, blockType, err = g.seekPreviousBlockFrom(pos) {
		if err != nil {
			return -1, err
		}
	}
	return pos, nil
}

func (g *Gouch) readAt(buf []byte, size int64, pos int64) (int64, error) {
	bytesReadSoFar := int64(0)
	bytesSkipped := int64(0)
	numBytesToRead := int64(size)
	readOffset := pos
	for numBytesToRead > 0 {
		var err error
		bytesTillNextBlock := BlockSize - (readOffset % BlockSize)
		if bytesTillNextBlock == BlockSize {
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
			return -1, err
		}
		readOffset += int64(n)
		bytesReadSoFar += int64(n)
		numBytesToRead -= int64(n)
		if int64(n) < bytesToReadThisPass {
			return bytesReadSoFar, nil
		}
	}
	return bytesReadSoFar + bytesSkipped, nil
}
