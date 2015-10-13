package gouch

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

//Gouch handler for reading a index file
type Gouch struct {
	file   *os.File
	pos    int64
	header *IndexHeader
	ops    Ops
}

//DocumentInfo Handler for capturing metadata
type DocumentInfo struct {
	ID    string `json:"id"`    // document identifier
	Key   string `json:"key"`   // emitted key
	Value string `json:"value"` // emitted value
}

//Open up index file with defined perms
func Open(filename string, options int) (*Gouch, error) {
	return OpenEx(filename, options, NewBaseOps())
}

//OpenEx opens index file and looks for valid header
func OpenEx(filename string, options int, ops Ops) (*Gouch, error) {
	gouch := Gouch{
		ops: ops,
	}

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
	}
	err = gouch.findLastHeader()
	if err != nil {
		return nil, err
	}
	return &gouch, nil
}

//Close clears up the open file handle
func (g *Gouch) Close() error {
	return g.ops.Close(g.file)
}

func (di *DocumentInfo) String() string {
	return fmt.Sprintf("ID: '%s' Key: '%s' Value: '%s' ", di.ID, di.Key, di.Value)
}

func lookupCallback(req *lookupRequest, key []byte, value []byte) error {
	if value == nil {
		return nil
	}

	context := req.callbackContext.(*lookupContext)

	docinfo := DocumentInfo{}
	if context.indexType == IndexTypeByID || context.indexType == IndexTypeMapR {
		sz := decodeRaw16(key[:2])
		docinfo.ID = string(key[len(key)-int(sz)+2:])
		docinfo.Key = string(key[2 : len(key)-int(sz)+2])
		docinfo.Value = string(value[5:])
	}

	if context.walkTreeCallback != nil {
		if context.indexType == IndexTypeLocalDocs {
			// note we pass the non-initialized docinfo so we can at least detect that its a leaf
			return context.walkTreeCallback(context.gouch, context.depth, &docinfo, key, 0, value, context.callbackContext)
		}
		return context.walkTreeCallback(context.gouch, context.depth, &docinfo, nil, 0, nil, context.callbackContext)
	} /*else if context.documentInfoCallback != nil {
		return context.documentInfoCallback(context.gouch, &docinfo, context.callbackContext)
	}*/

	return nil
}

func walkNodeCallback(req *lookupRequest, key []byte, value []byte) error {
	context := req.callbackContext.(*lookupContext)
	if value == nil {
		context.depth--
		return nil
	}
	//valueNodePointer := decodeNodePointer(value)
	valueNodePointer := &nodePointer{}
	valueNodePointer.subtreeSize = decodeRaw48(value)
	valueNodePointer.key = key
	size := decodeRaw16(value)
	valueNodePointer.reducedValue = value[14 : 14+size]
	valueNodePointer.reducedValue = value[14:]
	err := context.walkTreeCallback(context.gouch, context.depth, nil, key, valueNodePointer.subtreeSize, valueNodePointer.reducedValue, context.callbackContext)
	context.depth++
	return err
}

//DocumentInfoCallback callback for capturing metadata
type DocumentInfoCallback func(gouch *Gouch, documentInfo *DocumentInfo, userContext interface{}, limit int, w io.Writer) error

//AllDocuments dumps all documents based on startID and endID
func (g *Gouch) AllDocuments(startID, endID string, limit int, cb DocumentInfoCallback, userContext interface{}, w io.Writer) error {
	wtCallback := func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error {
		if documentInfo != nil {
			return cb(gouch, documentInfo, userContext, limit, w)
		}
		return nil
	}
	return g.WalkIDTree(startID, endID, wtCallback, userContext)
}

//WalkTreeCallback callback for traversing btree
type WalkTreeCallback func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error

//WalkIDTree traverses a btree based on startID and endID
func (g *Gouch) WalkIDTree(startID, endID string, wtcb WalkTreeCallback, userContext interface{}) error {

	if g.header.idBTreeState == nil {
		return nil
	}

	wtcb(g, 0, nil, nil, g.header.idBTreeState.subtreeSize, g.header.idBTreeState.reducedValue, userContext)

	lc := lookupContext{
		gouch:            g,
		walkTreeCallback: wtcb,
		callbackContext:  userContext,
		indexType:        IndexTypeByID,
	}

	keys := [][]byte{[]byte(startID)}
	if endID != "" {
		keys = append(keys, []byte(endID))
	}

	lr := lookupRequest{
		compare:         IDComparator,
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

//AllDocsMapReduce MapReduce tree dump
func (g *Gouch) AllDocsMapReduce(startID, endID string, limit int, mapR DocumentInfoCallback, userContext interface{}, w io.Writer) error {
	mapRCallback := func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error {
		if documentInfo != nil {
			return mapR(gouch, documentInfo, userContext, limit, w)
		}
		return nil
	}
	return g.WalkMapReduceTree(startID, endID, mapRCallback, userContext)
}

//WalkMapReduceTree MapReduce tree traversal
func (g *Gouch) WalkMapReduceTree(startID, endID string, mapR WalkTreeCallback, userContext interface{}) error {

	if len(g.header.viewStates) == 0 {
		return nil
	}

	for i := 0; i < len(g.header.viewStates); i++ {
		mapR(g, 0, nil, nil, g.header.viewStates[i].subtreeSize, g.header.viewStates[i].reducedValue, userContext)

		lc := lookupContext{
			gouch:            g,
			walkTreeCallback: mapR,
			callbackContext:  userContext,
			indexType:        IndexTypeMapR,
		}

		keys := [][]byte{[]byte(startID)}
		if endID != "" {
			keys = append(keys, []byte(endID))
		}

		lr := lookupRequest{
			compare:         IDComparator,
			keys:            keys,
			fetchCallback:   lookupCallback,
			nodeCallback:    walkNodeCallback,
			fold:            true,
			callbackContext: &lc,
		}

		err := g.btreeLookup(&lr, g.header.viewStates[i].pointer)
		if err != nil {
			return err
		}
	}
	return nil
}

//DefaultDocumentCallback sample document callback function
//TODO implement limit support
func DefaultDocumentCallback(g *Gouch, docInfo *DocumentInfo, userContext interface{}, limit int, w io.Writer) error {
	bytes, err := json.MarshalIndent(docInfo, "", " ")
	if err != nil {
		fmt.Println(err)
	} else {
		if userContext.(map[string]int)["count"] < limit {
			userContext.(map[string]int)["count"]++
			fmt.Println(string(bytes))
		} else {
			return nil
		}
	}
	return nil
}
