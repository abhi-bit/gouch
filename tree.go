package gouch

import (
	"fmt"
)

type DocumentInfoCallback func(gouch *Gouch, documentInfo *DocumentInfo, userContext interface{}) error

type WalkTreeCallback func(gouch *Gouch, depth int, documentInfo *DocumentInfo, key []byte, subTreeSize uint64, reducedValue []byte, userContext interface{}) error

type callback func(req *lookupRequest, key []byte, value []byte) error

type DocumentInfo struct {
	ID           string `json:"id"`
	Seq          uint64 `json:"seq"`
	Rev          uint64 `json:"rev"`
	RevMeta      []byte `json:"revMeta"`
	ContentMeta  uint8  `json:"contentMeta"`
	Deleted      bool   `json:"deleted"`
	Size         uint64 `json:"size"`
	bodyPosition uint64
}

type lookupContext struct {
	gouch                *Gouch
	documentInfoCallback DocumentInfoCallback
	walkTreeCallback     WalkTreeCallback
	indexType            int
	depth                int
	callbackContext      interface{}
}

type lookupRequest struct {
	gouch           *Gouch
	compare         btreeKeyComparator
	keys            [][]byte
	fold            bool
	inFold          bool
	fetchCallback   callback
	nodeCallback    callback
	callbackContext interface{}
}

func (g *Gouch) btreeLookupInner(req *lookupRequest, diskPos uint64, current, end int) error {
	nodeData, err := g.readCompressedDataChunkAt(int64(diskPos))
	if err != nil {
		return err
	}

	fmt.Printf("BtreeLookupInterior nodeData: %+v\n", nodeData)

	if nodeData[0] == BTREE_INTERIOR {
		kvIterator := newKVIterator(nodeData[1:])
		for k, v := kvIterator.Next(); k != nil && current < end; k, v = kvIterator.Next() {
			cmp := req.compare(k, req.keys[current])
			if cmp >= 0 {
				if req.fold {
					req.inFold = true
				}

				// Descend into the pointed to node.
				// with all keys < item key.
				lastItem := current + 1
				for lastItem < end && req.compare(k, req.keys[lastItem]) >= 0 {
					lastItem++
				}

				if req.nodeCallback != nil {
					err = req.nodeCallback(req, k, v)
					if err != nil {
						return err
					}
				}

				valNodePointer := decodeNodePointer(v)
				err = g.btreeLookupInner(req, valNodePointer.pointer, current, lastItem)
				if err != nil {
					return err
				}

				if !req.inFold {
					current = lastItem
				}
				if req.nodeCallback != nil {
					err = req.nodeCallback(req, nil, nil)
					if err != nil {
						return err
					}
				}
			}
		}
	} else if nodeData[0] == BTREE_LEAF {
		kvIterator := newKVIterator(nodeData[1:])
		for k, v := kvIterator.Next(); k != nil && current < end; k, v = kvIterator.Next() {
			cmp := req.compare(k, req.keys[current])
			if cmp >= 0 && req.fold && !req.inFold {
				req.inFold = true
			} else if req.inFold && (current+1) < end && req.compare(k, req.keys[current+1]) > 0 {
				//We've hit a key past the end of our range.
				req.inFold = false
				req.fold = false
				current = end
			}

			if cmp == 0 || (cmp > 0 && req.inFold) {
				// Found
				err = req.fetchCallback(req, k, v)
				if err != nil {
					return err
				}

				if !req.inFold {
					current++
				}
			}
		}
	}

	//Any remaining items are not found.
	for current < end && !req.fold {
		err = req.fetchCallback(req, req.keys[current], nil)
		current++
	}

	return nil
}

func (g *Gouch) btreeLookup(req *lookupRequest, rootPointer uint64) error {
	req.inFold = false
	fmt.Printf("Dumping inputs to btreeLookupInner: req: %+v rootPointer: %+v \n", req, rootPointer)
	return g.btreeLookupInner(req, rootPointer, 0, len(req.keys))
}
