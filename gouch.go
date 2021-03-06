package gouch

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

//Gouch handler for reading a index file
type Gouch struct {
	file          *os.File
	pos           int64
	header        *IndexHeader
	ops           Ops
	isFDAllocated bool
}

//DocumentInfo Handler for capturing metadata
type DocumentInfo struct {
	ID    []byte `json:"id"`    // document identifier
	Key   []byte `json:"key"`   // emitted key
	Value []byte `json:"value"` // emitted value
}

//GetFDStatus provides status of header FD
func (g *Gouch) GetFDStatus() bool {
	return g.isFDAllocated
}

//SetStatus assists in caching index header location
func (g *Gouch) SetStatus(state bool) {
	g.isFDAllocated = state
}

//DeepCopy copies one gouch struct into another
func (g *Gouch) DeepCopy() *Gouch {
	rv := &Gouch{g.file, g.pos, g.header, g.ops, g.isFDAllocated}
	return rv
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
	_, err = gouch.findLastHeader()
	if err != nil {
		return nil, err
	}
	return &gouch, nil
}

func (g *Gouch) GetHeaderCount() int64 {
	var headerCount int64
	var headerPos int64
	err := fmt.Errorf("headercount")
	headerPos, err = g.findLastHeader()
	for err == nil && headerPos > 4096 {
		headerCount = headerCount + 1
		g.pos = headerPos - 1
		fmt.Printf("header pos: %d\n", g.pos)
		headerPos, err = g.findLastHeader()
	}
	return headerCount
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

		docinfo.ID = key[int(sz)+2:]
		docinfo.Key = key[2 : int(sz)+2]
		docinfo.Value = value[5:]
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
	count := context.callbackContext.(map[string]int)["count"]

	if count > req.limit {
		return nil
	}

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
type DocumentInfoCallback func(gouch *Gouch, documentInfo *DocumentInfo, userContext interface{}, w io.Writer) error

//AllDocuments dumps all documents based on startID and endID
func (g *Gouch) AllDocuments(startID, endID string, cb DocumentInfoCallback, userContext interface{}, w io.Writer, limit int) error {
	wtCallback := func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error {
		if documentInfo != nil {
			return cb(gouch, documentInfo, userContext, w)
		}
		return nil
	}
	return g.WalkIDTree(startID, endID, wtCallback, userContext, limit)
}

//WalkTreeCallback callback for traversing btree
type WalkTreeCallback func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error

//WalkIDTree traverses a btree based on startID and endID
func (g *Gouch) WalkIDTree(startID, endID string, wtcb WalkTreeCallback, userContext interface{}, limit int) error {

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
func (g *Gouch) AllDocsMapReduce(startID, endID string, mapR DocumentInfoCallback, userContext interface{}, w io.Writer, limit int) error {

	mapRCallback := func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error {
		if documentInfo != nil {
			return mapR(gouch, documentInfo, userContext, w)
		}
		return nil
	}
	return g.WalkMapReduceTree(startID, endID, mapRCallback, userContext, limit)
}

//WalkMapReduceTree MapReduce tree traversal
func (g *Gouch) WalkMapReduceTree(startID, endID string, mapR WalkTreeCallback, userContext interface{}, limit int) error {

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
			limit:           limit,
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
func DefaultDocumentCallback(g *Gouch, docInfo *DocumentInfo, userContext interface{}, w io.Writer) error {
	//bytes, err := json.Marshal(docInfo)
	_, err := json.Marshal(docInfo)
	if err != nil {
		fmt.Println(err)
	} else {
		userContext.(map[string]int)["count"]++
		//fmt.Println(string(bytes))
	}
	return nil
}
