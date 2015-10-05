package gouch

import (
	"bytes"
)

type btreeKeyComparator func(a, b []byte) int

func IdComparator(a, b []byte) int {
	return bytes.Compare(a, b)
}

func SeqComparator(a, b []byte) int {
	aseq := decode_raw48(a)
	bseq := decode_raw48(b)

	if aseq < bseq {
		return -1
	} else if aseq == bseq {
		return 0
	}
	return 1
}

type seqList []uint64

func (s seqList) Len() int           { return len(s) }
func (s seqList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s seqList) Less(i, j int) bool { return s[i] < s[j] }

type idList [][]byte

func (idl idList) Len() int           { return len(idl) }
func (idl idList) Swap(i, j int)      { idl[i], idl[j] = idl[j], idl[i] }
func (idl idList) Less(i, j int) bool { return IdComparator(idl[i], idl[j]) < 0 }

type idAndValueList struct {
	ids  idList
	vals idList
}

func (idavl idAndValueList) Len() int { return idavl.ids.Len() }
func (idval idAndValueList) Swap(i, j int) {
	idval.ids.Swap(i, j)
	idval.vals[i], idval.vals[j] = idval.vals[j], idval.vals[i]
}
func (idval idAndValueList) Less(i, j int) bool { return idval.ids.Less(i, j) }

type partSeq struct {
	partID uint16
	seq    uint64
}
type partSeqList []partSeq

func (psl partSeqList) Len() int           { return len(psl) }
func (psl partSeqList) Swap(i, j int)      { psl[i], psl[j] = psl[j], psl[i] }
func (psl partSeqList) Less(i, j int) bool { return (psl[i].partID - psl[i].partID) < 0 }

type partVersion struct {
	partID         uint16
	numFailoverLog uint16
	failoverLog    failoverList
}
type partVersionList []partVersion

func (pvl partVersionList) Len() int           { return len(pvl) }
func (pvl partVersionList) Swap(i, j int)      { pvl[i], pvl[j] = pvl[j], pvl[i] }
func (pvl partVersionList) Less(i, j int) bool { return (pvl[i].partID - pvl[i].partID) < 0 }

type FailoverLog struct {
	uuid []byte
	seq  uint64
}
type failoverList []FailoverLog
