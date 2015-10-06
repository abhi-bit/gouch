package gouch

import (
	"bytes"
	"fmt"
)

const COUCH_BLOCK_SIZE = 4096
const COUCH_DISK_VERSION = 11
const COUCH_SNAPPY_THRESHOLD = 64
const MAX_DB_HEADER_SIZE = 1024

const (
	BTREE_INTERIOR        byte = 0
	BTREE_LEAF            byte = 1
	INDEX_TYPE_BY_ID      int  = 0
	INDEX_TYPE_BY_SEQ     int  = 1
	INDEX_TYPE_LOCAL_DOCS int  = 2
	KEY_VALUE_LEN         int  = 5
)

const ROOT_BASE_SIZE = 12

type node struct {
	pointers []*nodePointer
}

type nodePointer struct {
	key              []byte
	pointer          uint64
	subTreeSize      uint64
	reducedValueSize uint16
	reducedValue     []byte
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
	buf.Write(encode_raw16(np.reducedValueSize))
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
	if len(data) == 0 {
		return nil
	}

	np := nodePointer{}
	np.pointer = decode_raw48(data[0:6])
	np.subTreeSize = decode_raw48(data[6:12])
	np.reducedValueSize = decode_raw16(data[12:14])
	np.reducedValue = data[14:]
	return &np
}

func decodeByIdValue(docinfo *DocumentInfo, value []byte) {
	fmt.Printf("value dump: %+v\n", value)
	docinfo.Seq = decode_raw48(value[0:6])
	docinfo.Size = uint64(decode_raw32(value[6:10]))
	docinfo.Deleted, docinfo.bodyPosition = decode_raw_1_47_split(value[10:16])
	docinfo.Rev = decode_raw48(value[16:22])
	docinfo.ContentMeta = decode_raw08(value[22:23])
	docinfo.RevMeta = value[23:]
}

func decodeNodePointer(data []byte) *nodePointer {
	np := nodePointer{}
	fmt.Printf("data from decodeNodePointer: %+v\n", data)
	np.pointer = decode_raw48(data[0:6])
	np.subTreeSize = decode_raw48(data[6:12])
	np.reducedValueSize = decode_raw16(data[12:14])
	np.reducedValue = data[14:]
	fmt.Printf("nodepointer dump: %+v\n", np)
	return &np
}

func decodeBySeqValue(docinfo *DocumentInfo, value []byte) {
	idSize, docSize := decode_raw_12_28_split(value[0:5])
	docinfo.Size = uint64(docSize)
	docinfo.Deleted, docinfo.bodyPosition = decode_raw_1_47_split(value[5:12])
	docinfo.Rev = decode_raw48(value[11:17])
	docinfo.ContentMeta = decode_raw08(value[17:18])
	docinfo.ID = string(value[18 : 18+idSize])
	docinfo.RevMeta = value[18+idSize:]
}

func decodeKeyValue(nodeData []byte, bufPos int) ([]byte, []byte, int) {
	keyLength, valueLength := decode_raw_12_28_split(nodeData[bufPos : bufPos+5])
	keyStart := bufPos + 5
	keyEnd := keyStart + int(keyLength)
	key := nodeData[keyStart:keyEnd]
	valueStart := keyEnd
	valueEnd := valueStart + int(valueLength)
	value := nodeData[valueStart:valueEnd]
	return key, value, valueEnd
}

type KVIterator struct {
	data []byte
	pos  int
}

func newKVIterator(data []byte) *KVIterator {
	return &KVIterator{
		data: data,
	}
}

func (kvi *KVIterator) Next() ([]byte, []byte) {
	if kvi.pos < len(kvi.data) {
		key, value, end := decodeKeyValue(kvi.data, kvi.pos)
		kvi.pos = end
		return key, value
	}
	return nil, nil
}
