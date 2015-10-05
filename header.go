package gouch

import (
	"fmt"
	"sort"
)

const LATEST_INDEX_HEADER_VERSION = 2

type IndexStateTransition struct {
	active      *SortedList
	passive     *SortedList
	unindexable *SortedList
}

type indexHeader struct {
	version            uint8
	signature          []byte
	numViews           uint8
	numPartitions      uint16
	activeBitmask      *Bitmap
	passiveBitmask     *Bitmap
	cleanupBitmask     *Bitmap
	seqs               seqList
	idBTreeState       *nodePointer
	viewStates         *nodePointer
	hasReplica         int
	replicasOnTransfer seqList
	pendingTransaction IndexStateTransition
	unindexableSeqs    partSeqList
	partVersions       partVersionList
}

func DecodeIndexHeader(bytes []byte) *indexHeader {
	if len(bytes) <= 16 {
		fmt.Errorf("Corrupt header\n")
		return nil
	}
	var data []byte
	arrayIndex := 0
	//SnappyDecode(bytes[3:], data)

	h := indexHeader{}

	h.signature = bytes[0:2]
	h.version = decode_raw08(data[arrayIndex : arrayIndex+1])
	arrayIndex++
	h.numPartitions = decode_raw16(data[arrayIndex+1 : arrayIndex+3])
	arrayIndex += 3

	// BITMAP_SIZE == 128
	arrayIndex++
	h.activeBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BITMAP_SIZE]}
	arrayIndex += BITMAP_SIZE
	arrayIndex++
	h.passiveBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BITMAP_SIZE]}
	arrayIndex += BITMAP_SIZE
	arrayIndex++
	h.cleanupBitmask = &Bitmap{data: data[arrayIndex : arrayIndex+BITMAP_SIZE]}
	arrayIndex += BITMAP_SIZE

	arrayIndex++
	numSeqs := decode_raw16(data[arrayIndex : arrayIndex+2])
	arrayIndex += 2

	for i := 0; i < int(numSeqs); i++ {
		arrayIndex++
		ps := partSeq{
			partID: decode_raw16(data[arrayIndex : arrayIndex+2]),
			seq:    decode_raw48(data[arrayIndex+3 : arrayIndex+9]),
		}
		arrayIndex += 9
		h.unindexableSeqs = append(h.unindexableSeqs, ps)
	}
	sort.Sort(h.unindexableSeqs)

	if h.version >= 2 {
		arrayIndex++
		numPartVersions := decode_raw16(data[arrayIndex : arrayIndex+2])

		for i := 0; i < int(numPartVersions); i++ {
			arrayIndex++
			pver := partVersion{
				partID:         decode_raw16(data[arrayIndex : arrayIndex+2]),
				numFailoverLog: decode_raw16(data[arrayIndex+3 : arrayIndex+5]),
			}
			arrayIndex += 5
			for j := 0; j < int(pver.numFailoverLog); j++ {
				arrayIndex++
				fl := FailoverLog{
					uuid: data[arrayIndex : arrayIndex+8],
					seq:  decode_raw64(data[arrayIndex+9 : arrayIndex+17]),
				}
				arrayIndex += 17
				pver.failoverLog = append(pver.failoverLog, fl)
			}
			h.partVersions = append(h.partVersions, pver)
		}
		sort.Sort(h.partVersions)
	}

	return &h
}

/*func (g *Gouch) readHeaderAt(pos int64) (*header, error) {
	chunk, err := g.readChunkAt(pos, true)
	if err != nil {
		return nil, err
	}
}*/

func (g *Gouch) findLastHeader() error {
	pos := g.pos
	var h *indexHeader
	var err error
	for h == nil && err != nil {
		_, err = g.seekLastHeaderBlockFrom(pos)
		if err != nil {
			return err
		}
	}
	g.header = h
	return nil
}
