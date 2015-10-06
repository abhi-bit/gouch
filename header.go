package gouch

import (
	"fmt"
	"reflect"
	"sort"
)

const LATEST_INDEX_HEADER_VERSION = 2

type IndexStateTransition struct {
	active      seqList
	passive     seqList
	unindexable seqList
}

type nodePointerList []*nodePointer

type indexHeader struct {
	version            uint8
	signature          []byte
	numViews           uint8
	numPartitions      uint16
	activeBitmask      *Bitmap
	passiveBitmask     *Bitmap
	cleanupBitmask     *Bitmap
	seqs               partSeqList
	idBTreeState       *nodePointer
	viewStates         nodePointerList
	hasReplica         int
	replicasOnTransfer seqList
	pendingTransaction IndexStateTransition
	unindexableSeqs    partSeqList
	partVersions       partVersionList
}

func DecodeIndexHeader(bytes []byte) *indexHeader {
	if len(bytes) <= 16 {
		fmt.Printf("Corrupt header len: %+v\n", len(bytes))
		return nil
	}
	fmt.Printf("Sane header len: %+v\n", len(bytes))
	var data []byte
	arrayIndex := 0

	data, err := SnappyDecode(nil, bytes[16:])
	fmt.Printf("Input array size: %+v\n", reflect.TypeOf(bytes).Size())
	if err != nil {
		fmt.Printf("error in snappy decode: %+v\n", err)
		return nil
	}

	h := indexHeader{}

	h.signature = bytes[0:15]
	fmt.Printf("Data array: %+v\n", data[0:10])
	h.version = decode_raw08(data[arrayIndex : arrayIndex+1])
	arrayIndex += 1
	h.numPartitions = decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2
	fmt.Printf("Version: %+v Numpartitions: %+v\n", h.version, h.numPartitions)

	// BITMAP_SIZE == 128
	h.activeBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BITMAP_SIZE]}
	arrayIndex += BITMAP_SIZE
	//fmt.Printf("active bitmask dump: %+v\n", h.activeBitmask.Dump())
	h.passiveBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BITMAP_SIZE]}
	arrayIndex += BITMAP_SIZE
	//fmt.Printf("passive bitmask dump: %+v\n", h.passiveBitmask.Dump())
	h.cleanupBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BITMAP_SIZE]}
	arrayIndex += BITMAP_SIZE
	//fmt.Printf("cleanup bitmask dump: %+v\n", h.cleanupBitmask.Dump())

	numSeqs := decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2
	fmt.Printf("numSeqs: %+v\n", numSeqs)

	for i := 0; i < int(numSeqs); i++ {
		ps := partSeq{
			partID: decode_raw16(data[arrayIndex : arrayIndex+2]),
			seq:    decode_raw48(data[arrayIndex+2 : arrayIndex+8]),
		}
		//fmt.Printf("partID: %+v seq: %+v\n", ps.partID, ps.seq)
		arrayIndex += 8
		h.seqs = append(h.seqs, ps)
	}
	sort.Sort(h.seqs)
	//fmt.Printf("h.seqs: %+v\n", h.seqs)

	sz := decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2
	//fmt.Printf("Size: %d\n", sz)

	h.idBTreeState = decodeRootNodePointer(data[arrayIndex : arrayIndex+int(sz)])
	arrayIndex += int(sz)

	h.numViews = decode_raw08(data[arrayIndex : arrayIndex+1])
	arrayIndex += 1
	fmt.Printf("numViews: %+v\n", h.numViews)

	h.viewStates = make([]*nodePointer, int(h.numViews))

	for i := 0; i < int(h.numViews); i++ {
		sz = decode_raw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.viewStates[i] = decodeRootNodePointer(data[arrayIndex : arrayIndex+int(sz)])
		arrayIndex += int(sz)
	}

	h.hasReplica = int(decode_raw08(data[arrayIndex : arrayIndex+1]))
	arrayIndex += 1

	sz = decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decode_raw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.replicasOnTransfer = append(h.replicasOnTransfer, uint64(partID))
	}
	sort.Sort(h.replicasOnTransfer)

	sz = decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decode_raw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.pendingTransaction.active = append(h.pendingTransaction.active, uint64(partID))
	}
	sort.Sort(h.pendingTransaction.active)

	sz = decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decode_raw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.pendingTransaction.passive = append(h.pendingTransaction.passive, uint64(partID))
	}
	sort.Sort(h.pendingTransaction.passive)

	sz = decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decode_raw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.pendingTransaction.unindexable = append(h.pendingTransaction.unindexable, uint64(partID))
	}
	sort.Sort(h.pendingTransaction.unindexable)

	numSeqs = decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(numSeqs); i++ {
		pSeq := partSeq{}
		pSeq.partID = decode_raw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		pSeq.seq = decode_raw64(data[arrayIndex : arrayIndex+6])
		arrayIndex += 6

		h.unindexableSeqs = append(h.unindexableSeqs, pSeq)
	}
	sort.Sort(h.unindexableSeqs)

	if h.version >= 2 {
		numPartVersions := decode_raw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2

		for i := 0; i < int(numPartVersions); i++ {
			pver := partVersion{}
			pver.partID = decode_raw16(data[arrayIndex : arrayIndex+2])
			arrayIndex += 2
			pver.numFailoverLog = decode_raw16(data[arrayIndex : arrayIndex+2])
			arrayIndex += 2

			for j := 0; j < int(pver.numFailoverLog); j++ {
				fl := FailoverLog{}
				fl.uuid = data[arrayIndex : arrayIndex+8]
				arrayIndex += 8
				fl.seq = decode_raw64(data[arrayIndex : arrayIndex+8])
				arrayIndex += 8

				pver.failoverLog = append(pver.failoverLog, fl)
			}
			h.partVersions = append(h.partVersions, pver)
		}
		sort.Sort(h.partVersions)
	}
	//fmt.Printf("Header dump: %+v\n", h)
	return &h
}

func (g *Gouch) readHeaderAt(pos int64) (*indexHeader, error) {
	chunk, err := g.readChunkAt(pos, true)
	if err != nil {
		return nil, err
	}
	header := DecodeIndexHeader(chunk)
	return header, nil
}

func (g *Gouch) findLastHeader() error {
	pos := g.pos
	var h *indexHeader
	var err error = fmt.Errorf("start")
	var headerPos int64
	for h == nil && err != nil {
		headerPos, err = g.seekLastHeaderBlockFrom(pos)
		if err != nil {
			return err
		}
		h, err = g.readHeaderAt(headerPos)
		if err != nil {
			pos = headerPos - 1
		}
	}
	g.header = h
	return nil
}
