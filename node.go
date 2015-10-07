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
	fmt.Printf("Input byte array: %+v\n", len(data))
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
		key, value, end := decodeKeyValue(nodeData, bufPos)
		docinfo := DocumentInfo{}
		if indexType == IndexTypeByID {
			docinfo.ID = string(key)
			decodeByIDValue(&docinfo, value)
		} else if indexType == IndexTypeBySeq {
			docinfo.Seq = decodeRaw48(key)
			decodeBySeqValue(&docinfo, value)
		}

		resultNode.documents = append(resultNode.documents, &docinfo)
		bufPos = end
	}
	return resultNode, nil
}

func decodeByIDValue(docinfo *DocumentInfo, value []byte) {
	docinfo.Seq = decodeRaw48(value[0:6])
	docinfo.Size = uint64(decodeRaw32(value[6:10]))
	//docinfo.Deleted, docinfo.bodyPosition = decode_raw_1_47_split(value[10:16])
	//docinfo.Rev = decode_raw48(value[16:22])
	//docinfo.ContentMeta = decode_raw08(value[22:23])
	//docinfo.RevMeta = value[23:]
}

func (d DocumentInfo) encodeByID() []byte {
	buf := new(bytes.Buffer)
	buf.Write(encodeRaw48(d.Seq))
	buf.Write(encodeRaw32(d.Size))
	buf.Write(encodeRaw1_47Split(d.Deleted, d.bodyPosition))
	buf.Write(encodeRaw48(d.Rev))
	buf.Write(encodeRaw08(d.ContentMeta))
	buf.Write(d.RevMeta)
	return buf.Bytes()
}

func decodeBySeqValue(docinfo *DocumentInfo, value []byte) {
	idSize, docSize := decodeRaw12_28Split(value[0:5])
	docinfo.Size = uint64(docSize)
	docinfo.Deleted, docinfo.bodyPosition = decodeRaw1_47Split(value[5:12])
	docinfo.Rev = decodeRaw48(value[11:17])
	docinfo.ContentMeta = decodeRaw08(value[17:18])
	docinfo.ID = string(value[18 : 18+idSize])
	docinfo.RevMeta = value[18+idSize:]
}

func (d DocumentInfo) encodeBySeq() []byte {
	buf := new(bytes.Buffer)
	buf.Write(encodeRaw12_28Split(uint32(len(d.ID)), uint32(d.Size)))
	buf.Write(encodeRaw1_47Split(d.Deleted, d.bodyPosition))
	buf.Write(encodeRaw48(d.Rev))
	buf.Write(encodeRaw08(d.ContentMeta))
	buf.Write([]byte(d.ID))
	buf.Write(d.RevMeta)
	return buf.Bytes()
}

func decodeKeyValue(nodeData []byte, bufPos int) ([]byte, []byte, int) {
	keyLength, valueLength := decodeRaw12_28Split(nodeData[bufPos : bufPos+5])
	keyStart := bufPos + 5
	keyEnd := keyStart + int(keyLength)
	key := nodeData[keyStart:keyEnd]
	valueStart := keyEnd
	valueEnd := valueStart + int(valueLength)
	value := nodeData[valueStart:valueEnd]
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
