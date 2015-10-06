package gouch

import (
	"fmt"
	"log"
	"os"
)

type Gouch struct {
	file   *os.File
	pos    int64
	header *indexHeader
	ops    Ops
}

type DocumentInfo struct {
	ID           string `json:"id"`          // document identifier
	Seq          uint64 `json:"seq"`         // sequence number in database
	Rev          uint64 `json:"rev"`         // revision number of document
	RevMeta      []byte `json:"revMeta"`     // additional revision meta-data (uninterpreted by Gouchstore)
	ContentMeta  uint8  `json:"contentMeta"` // content meta-data flags
	Deleted      bool   `json:"deleted"`     // is the revision deleted?
	Size         uint64 `json:"size"`        // size of document data in bytes
	bodyPosition uint64 // byte offset of document body in file
}

func Open(filename string, options int) (*Gouch, error) {
	return OpenEx(filename, options, NewBaseOps())
}

func OpenEx(filename string, options int, ops Ops) (*Gouch, error) {
	gouch := Gouch{
		ops: ops,
	}

	log.Printf("Trying to read file: %+v\n", filename)
	file, err := gouch.ops.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	gouch.file = file

	gouch.pos, err = gouch.ops.GotoEOF(gouch.file)
	if err != nil {
		fmt.Errorf("Failed while reading file from the end. file: %v\n", filename)
		return nil, err
	}

	if gouch.pos == 0 {
		fmt.Errorf("Empty file: %v\n", filename)
		return nil, err
	} else {
		err = gouch.findLastHeader()
		if err != nil {
			return nil, err
		}
	}
	return &gouch, nil
}

func lookupCallback(req *lookupRequest, key []byte, value []byte) error {
	if value == nil {
		return nil
	}

	context := req.callbackContext.(*lookupContext)

	docinfo := DocumentInfo{}
	if context.indexType == INDEX_TYPE_BY_ID {
		docinfo.ID = string(key)
		decodeByIdValue(&docinfo, value)
	} else if context.indexType == INDEX_TYPE_BY_SEQ {
		docinfo.Seq = decode_raw48(key)
		decodeBySeqValue(&docinfo, value)
	}

	if context.walkTreeCallback != nil {
		if context.indexType == INDEX_TYPE_LOCAL_DOCS {
			// note we pass the non-initialized docinfo so we can at least detect that its a leaf
			return context.walkTreeCallback(context.gouch, context.depth, &docinfo, key, 0, value, context.callbackContext)
		} else {
			return context.walkTreeCallback(context.gouch, context.depth, &docinfo, nil, 0, nil, context.callbackContext)
		}
	} else if context.documentInfoCallback != nil {
		return context.documentInfoCallback(context.gouch, &docinfo, context.callbackContext)
	}

	return nil
}

func walkNodeCallback(req *lookupRequest, key []byte, value []byte) error {
	context := req.callbackContext.(*lookupContext)
	if value == nil {
		context.depth--
		return nil
	} else {
		valueNodePointer := decodeNodePointer(value)
		valueNodePointer.key = key
		err := context.walkTreeCallback(context.gouch, context.depth, nil, key, valueNodePointer.subtreeSize, valueNodePointer.reducedValue, context.callbackContext)
		context.depth++
		return err
	}
}

type DocumentInfoCallback func(gouch *Gouch, documentInfo *DocumentInfo, userContext interface{}) error

func (g *Gouch) AllDocuments(startId, endId string, cb DocumentInfoCallback, userContext interface{}) error {
	wtCallback := func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error {
		if documentInfo != nil {
			return cb(gouch, documentInfo, userContext)
		}
		return nil
	}
	return g.WalkIdTree(startId, endId, wtCallback, userContext)
}

type WalkTreeCallback func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error

func (g *Gouch) WalkIdTree(startId, endId string, wtcb WalkTreeCallback, userContext interface{}) error {

	if g.header.idBTreeState == nil {
		return nil
	}

	wtcb(g, 0, nil, nil, g.header.idBTreeState.subtreeSize, g.header.idBTreeState.reducedValue, userContext)

	lc := lookupContext{
		gouch:            g,
		walkTreeCallback: wtcb,
		callbackContext:  userContext,
		indexType:        INDEX_TYPE_BY_ID,
	}

	keys := [][]byte{[]byte(startId)}
	if endId != "" {
		keys = append(keys, []byte(endId))
	}

	lr := lookupRequest{
		compare:         IdComparator,
		keys:            keys,
		fetchCallback:   lookupCallback,
		nodeCallback:    walkNodeCallback,
		fold:            true,
		callbackContext: &lc,
	}

	err := g.btreeLookup(&lr, g.header.idBTreeState.pointer)
	if err != nil {
		return err
	}

	return nil
}
