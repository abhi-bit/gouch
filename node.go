package gouch

import (
	"bytes"
	"fmt"
)

//BtreeInterior Btree interior node marker or KP node
const BtreeInterior byte = 0

//BtreeLeaf leaf node of BTree or KV node
const BtreeLeaf byte = 1

//IndexTypeByID By-ID Btree
const IndexTypeByID int = 0

//IndexTypeBySeq By_Seq Btree
const IndexTypeBySeq int = 1

//IndexTypeLocalDocs Local_Docs
const IndexTypeLocalDocs int = 2

//IndexTypeMapR map reduce index node type
const IndexTypeMapR int = 3

//RootBaseSize marker
const RootBaseSize int = 12

//KeyValueLen marker
const KeyValueLen int = 5

type node struct {
	// interior nodes will have this
	pointers []*nodePointer
	// leaf nodes will have this
	documents []*DocumentInfo
}

func (n *node) String() string {
	var rv string
	if n.pointers != nil {
		rv = "Interior Node: [\n"
		for i, p := range n.pointers {
			if i != 0 {
				rv += ",\n"
			}
			rv += fmt.Sprintf("%v", p)
		}
		rv += "\n]\n"
	} else {
		rv = "Leaf Node: [\n"
		for i, d := range n.documents {
			if i != 0 {
				rv += ",\n"
			}
			rv += fmt.Sprintf("%v", d)
		}
		rv += "\n]\n"
	}
	return rv
}

func newInteriorNode() *node {
	return &node{
		pointers: make([]*nodePointer, 0),
	}
}

func newLeafNode() *node {
	return &node{
		documents: make([]*DocumentInfo, 0),
	}
}

type sizedBuf struct {
}

type nodePointer struct {
	key          []byte
	pointer      uint64
	reducedValue []byte
	subtreeSize  uint64
}

func (np *nodePointer) encodeRoot() []byte {
	buf := new(bytes.Buffer)
	buf.Write(encodeRaw48(np.pointer))
	buf.Write(encodeRaw48(np.subtreeSize))
	buf.Write(np.reducedValue)
	return buf.Bytes()
}

func (np *nodePointer) encode() []byte {
	buf := new(bytes.Buffer)
	buf.Write(encodeRaw48(np.pointer))
	buf.Write(encodeRaw48(np.subtreeSize))
	buf.Write(encodeRaw16(uint16(len(np.reducedValue))))
	buf.Write(np.reducedValue)
	return buf.Bytes()
}

func decodeRootNodePointer(data []byte) *nodePointer {
	n := nodePointer{}
	n.pointer = decodeRaw48(data[0:6])
	n.subtreeSize = decodeRaw48(data[6:12])
	n.reducedValue = data[RootBaseSize:]
	return &n
}

func decodeNodePointer(data []byte) *nodePointer {
	n := nodePointer{}
	n.pointer = decodeRaw48(data[0:6])
	n.subtreeSize = decodeRaw48(data[6:12])
	reduceValueSize := decodeRaw16(data[12:14])
	n.reducedValue = data[14 : 14+reduceValueSize]
	return &n
}

func decodeInteriorBtreeNode(nodeData []byte, indexType int) (*node, error) {
	bufPos := 1
	resultNode := newInteriorNode()
	for bufPos < len(nodeData) {
		key, value, end := decodeKeyValue(nodeData, bufPos)
		valueNodePointer := decodeNodePointer(value)
		valueNodePointer.key = key
		resultNode.pointers = append(resultNode.pointers, valueNodePointer)
		bufPos = end
	}
	return resultNode, nil
}

func decodeLeafBtreeNode(nodeData []byte, indexType int) (*node, error) {
	bufPos := 1
	resultNode := newLeafNode()
	for bufPos < len(nodeData) {
		key, _, end := decodeKeyValue(nodeData, bufPos)
		docinfo := DocumentInfo{}
		if indexType == IndexTypeByID {
			docinfo.ID = string(key)
		}

		resultNode.documents = append(resultNode.documents, &docinfo)
		bufPos = end
	}
	return resultNode, nil
}

func decodeKeyValue(nodeData []byte, bufPos int) ([]byte, []byte, int) {
	keyLength, valueLength := decodeRaw12_28Split(nodeData[bufPos : bufPos+5])
	keyStart := bufPos + 5
	keyEnd := keyStart + int(keyLength)
	key := nodeData[keyStart:keyEnd]
	valueStart := keyEnd
	valueEnd := valueStart + int(valueLength)
	value := nodeData[valueStart:valueEnd]
	//TODO why we need byte offset of 2?
	return key, value, valueEnd
}

func encodeKeyValue(key, value []byte) []byte {
	buf := new(bytes.Buffer)
	keyLength := len(key)
	valueLength := len(value)
	buf.Write(encodeRaw12_28Split(uint32(keyLength), uint32(valueLength)))
	buf.Write(key)
	buf.Write(value)
	return buf.Bytes()
}

type keyValueIterator struct {
	data []byte
	pos  int
}

func newKeyValueIterator(data []byte) *keyValueIterator {
	return &keyValueIterator{
		data: data,
	}
}

func (kvi *keyValueIterator) Next() ([]byte, []byte) {
	if kvi.pos < len(kvi.data) {
		key, value, end := decodeKeyValue(kvi.data, kvi.pos)
		kvi.pos = end
		return key, value
	}
	return nil, nil
}

func (np *nodePointer) String() string {
	if np.key == nil {
		return fmt.Sprintf("Root Pointer: %d Subtree Size: %d ReduceValue: % x", np.pointer, np.subtreeSize, np.reducedValue)
	}
	return fmt.Sprintf("Key: '%s' (% x) Pointer: %d Subtree Size: %d ReduceValue: % x", np.key, np.key, np.pointer, np.subtreeSize, np.reducedValue)
}
