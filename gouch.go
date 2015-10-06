package gouch

import (
	"fmt"
	"os"
)

type Gouch struct {
	file   *os.File
	pos    int64
	header *indexHeader
	ops    Ops
}

func Open(filename string, options int) (*Gouch, error) {
	return OpenEx(filename, options, NewBaseOps())
}

func OpenEx(filename string, options int, ops Ops) (*Gouch, error) {
	gouch := Gouch{
		ops: ops,
	}

	fmt.Printf("Trying to read file: %+v\n", filename)
	file, err := gouch.ops.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Errorf("Failed to open file: %v\n", filename)
		return nil, err
	}
	gouch.file = file

	fmt.Println("Trying to read from EOF")
	gouch.pos, err = gouch.ops.GotoEOF(gouch.file)
	if err != nil {
		fmt.Errorf("Failed while reading file from the end. file: %v\n", filename)
		return nil, err
	}

	if gouch.pos == 0 {
		fmt.Errorf("Empty file: %v\n", filename)
		return nil, err
	} else {
		fmt.Println("Trying to read last header")
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
	fmt.Printf("Key: %+v value: %+v req: %+v\n", string(key), string(value), req)
	if value == nil {
		fmt.Println("ABHI: value == nil")
		context.depth--
		return nil
	} else {
		valueNodePointer := decodeNodePointer(value)
		fmt.Printf("decodeNodePointer valueNodePointer: %+v\n", valueNodePointer)
		valueNodePointer.key = key
		err := context.walkTreeCallback(context.gouch, context.depth, nil, key, valueNodePointer.subTreeSize, valueNodePointer.reducedValue, context.callbackContext)
		context.depth++
		return err
	}
}

func (g *Gouch) AllDocuments(startId, endId string, cb DocumentInfoCallback, userContext interface{}) error {
	wtCallback := func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error {
		if documentInfo != nil {
			return cb(gouch, documentInfo, userContext)
		}
		return nil
	}
	return g.WalkIdTree(startId, endId, wtCallback, userContext)
}

func (g *Gouch) WalkIdTree(startId, endId string, wtcb WalkTreeCallback, userContext interface{}) error {

	if g.header.idBTreeState == nil {
		return nil
	}

	fmt.Printf("ABHI: idBtreeState: %+v\n", g.header.idBTreeState)
	wtcb(g, 0, nil, nil, g.header.idBTreeState.subTreeSize, g.header.idBTreeState.reducedValue, userContext)
	fmt.Printf("Gouch handle: %+v\n", g)

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

	fmt.Printf("idBTreeState: %+v\n", g.header.idBTreeState)
	err := g.btreeLookup(&lr, g.header.idBTreeState.pointer)
	if err != nil {
		return err
	}

	return nil
}
