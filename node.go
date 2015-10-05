package gouch

import (
	"bytes"
)

const COUCH_BLOCK_SIZE = 4096
const COUCH_DISK_VERSION = 11
const COUCH_SNAPPY_THRESHOLD = 64
const MAX_DB_HEADER_SIZE = 1024

const BTREE_INTERIOR byte = 0
const BTREE_LEAF byte = 1
const INDEX_TYPE_BY_ID int = 0
const INDEX_TYPE_BY_SEQ int = 1
const KEY_VALUE_LEN int = 5

const ROOT_BASE_SIZE = 12

type node struct {
	pointers []*nodePointer
}

type nodePointer struct {
	key          []byte
	pointer      uint64
	reducedValue []byte
	subTreeSize  uint64
}

func newInteriorNode() *node {
	return &node{
		pointers: make([]*nodePointer, 0),
	}
}

func (np *nodePointer) encodeRoot() []byte {
	buf := new(bytes.Buffer)
	buf.Write(encode_raw48(np.pointer))
	buf.Write(encode_raw48(np.subTreeSize))
	buf.Write(np.reducedValue)
	return buf.Bytes()
}

func (np *nodePointer) encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write(encode_raw48(np.pointer))
	buf.Write(encode_raw48(np.subTreeSize))
	buf.Write(encode_raw16(uint16(len(np.reducedValue))))
	buf.Write(np.reducedValue)
	return buf.Bytes()
}

func decodeRootNodePointer(data []byte) *nodePointer {
	np := nodePointer{}
	np.pointer = decode_raw48(data[0:6])
	np.subTreeSize = decode_raw48(data[6:12])
	np.reducedValue = data[ROOT_BASE_SIZE:]
	return &np
}

func decodeNodePointer(data []byte) *nodePointer {
	np := nodePointer{}
	np.pointer = decode_raw48(data[0:6])
	np.subTreeSize = decode_raw48(data[6:12])
	reduceValueSize := decode_raw16(data[12:14])
	np.reducedValue = data[14 : 14+reduceValueSize]
	return &np
}
