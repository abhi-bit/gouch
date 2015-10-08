package gouch

import (
	"fmt"
	"sort"
)

//IndexStateTransition capture the state of different partitions
type IndexStateTransition struct {
	active      seqList
	passive     seqList
	unindexable seqList
}

type nodePointerList []*nodePointer

//IndexHeader main header of index file
type IndexHeader struct {
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

func (g *Gouch) findLastHeader() error {
	pos := g.pos
	var h *IndexHeader
	err := fmt.Errorf("start")
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

func (g *Gouch) readHeaderAt(pos int64) (*IndexHeader, error) {
	chunk, err := g.readChunkAt(pos, true)
	if err != nil {
		return nil, err
	}
	header := DecodeIndexHeader(chunk)
	return header, nil
}

//DecodeIndexHeader decodes the main index header
func DecodeIndexHeader(bytes []byte) *IndexHeader {
	if len(bytes) <= 16 {
		return nil
	}
	var data []byte
	arrayIndex := 0

	data, err := SnappyDecode(nil, bytes[16:])
	if err != nil {
		return nil
	}

	h := IndexHeader{}

	h.signature = bytes[0:15]
	h.version = decodeRaw08(data[arrayIndex : arrayIndex+1])
	arrayIndex++
	h.numPartitions = decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	// BitmapSize == 128
	h.activeBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BitmapSize]}
	arrayIndex += BitmapSize
	h.passiveBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BitmapSize]}
	arrayIndex += BitmapSize
	h.cleanupBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BitmapSize]}
	arrayIndex += BitmapSize

	numSeqs := decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(numSeqs); i++ {
		ps := partSeq{}
		ps.partID = decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		ps.seq = decodeRaw48(data[arrayIndex : arrayIndex+6])
		arrayIndex += 6
		h.seqs = append(h.seqs, ps)
	}
	sort.Sort(h.seqs)

	sz := decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	h.idBTreeState = decodeRootNodePointer(data[arrayIndex : arrayIndex+int(sz)])
	arrayIndex += int(sz)

	h.numViews = decodeRaw08(data[arrayIndex : arrayIndex+1])
	arrayIndex++

	h.viewStates = make([]*nodePointer, int(h.numViews))

	for i := 0; i < int(h.numViews); i++ {
		sz = decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.viewStates[i] = decodeRootNodePointer(data[arrayIndex : arrayIndex+int(sz)])
		arrayIndex += int(sz)
	}

	h.hasReplica = int(decodeRaw08(data[arrayIndex : arrayIndex+1]))
	arrayIndex++

	sz = decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.replicasOnTransfer = append(h.replicasOnTransfer, uint64(partID))
	}
	sort.Sort(h.replicasOnTransfer)

	sz = decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.pendingTransaction.active = append(h.pendingTransaction.active, uint64(partID))
	}
	sort.Sort(h.pendingTransaction.active)

	sz = decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.pendingTransaction.passive = append(h.pendingTransaction.passive, uint64(partID))
	}
	sort.Sort(h.pendingTransaction.passive)

	sz = decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(sz); i++ {
		partID := decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		h.pendingTransaction.unindexable = append(h.pendingTransaction.unindexable, uint64(partID))
	}
	sort.Sort(h.pendingTransaction.unindexable)

	numSeqs = decodeRaw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(numSeqs); i++ {
		pSeq := partSeq{}
		pSeq.partID = decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2
		pSeq.seq = decodeRaw64(data[arrayIndex : arrayIndex+6])
		arrayIndex += 6

		h.unindexableSeqs = append(h.unindexableSeqs, pSeq)
	}
	sort.Sort(h.unindexableSeqs)

	if h.version >= 2 {
		numPartVersions := decodeRaw16(data[arrayIndex : arrayIndex+2])
		arrayIndex += 2

		for i := 0; i < int(numPartVersions); i++ {
			pver := partVersion{}
			pver.partID = decodeRaw16(data[arrayIndex : arrayIndex+2])
			arrayIndex += 2
			pver.numFailoverLog = decodeRaw16(data[arrayIndex : arrayIndex+2])
			arrayIndex += 2

			for j := 0; j < int(pver.numFailoverLog); j++ {
				fl := FailoverLog{}
				fl.uuid = data[arrayIndex : arrayIndex+8]
				arrayIndex += 8
				fl.seq = decodeRaw64(data[arrayIndex : arrayIndex+8])
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
