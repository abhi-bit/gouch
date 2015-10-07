package gouch

import (
	"container/list"
	"encoding/binary"
	"fmt"
)

//Item within SortedList
type Item struct {
	Data []byte
}

//CompareFunc comparator for given sorted list
type CompareFunc func(i1, i2 *Item) int

//SortedList data struct
type SortedList struct {
	List    *list.List
	Compare CompareFunc
}

//SortedListCreate creates a sortedList
func SortedListCreate(Compare CompareFunc) *SortedList {
	return &SortedList{
		List:    list.New(),
		Compare: Compare,
	}
}

//SortedListAdd adds an item to existing sorted list
func (s *SortedList) SortedListAdd(item *Item) {
	node := s.List.Front()
	var prev *list.Element

	// Initializing to a non-zero value
	cmp := -2
	for node != nil {
		if val, ok := node.Value.(*Item); ok {
			cmp = s.Compare(val, item)
			if cmp >= 0 {
				break
			}
			prev = node
			node = node.Next()
		}
	}

	if prev != nil {
		s.List.InsertAfter(item, prev)
	} else if prev == nil && node == nil {
		s.List.PushBack(item)
	}

	if cmp == 0 {
		s.List.InsertAfter(item, node)
	}
}

//SortedListGet grabs an item from SortedList
func (s *SortedList) SortedListGet(item *Item) *Item {
	node := s.List.Front()

	// Initializing to a non-zero value
	cmp := -2

	for node != nil {
		cmp = s.Compare(node.Value.(*Item), item)
		if cmp == 0 {
			return node.Value.(*Item)
		} else if cmp > 0 {
			return nil
		} else {
			node = node.Next()
		}
	}
	return nil
}

//SortedListRemove removes an item from SortedList
func (s *SortedList) SortedListRemove(item *Item) {
	node := s.List.Front()

	// Initializing to a non-zero value
	cmp := -2

	for node != nil {
		cmp = s.Compare(node.Value.(*Item), item)
		if cmp == 0 {
			s.List.Remove(node)
			return
		} else if cmp >= 0 {
			return
		} else {
			node = node.Next()
		}
	}
	return
}

//SortedListDump dumps the SortedList data struct
func (s *SortedList) SortedListDump() {
	for e := s.List.Front(); e != nil; e = e.Next() {
		fmt.Println(binary.BigEndian.Uint32(e.Value.(*Item).Data))
	}
}

//SortedListSize dumps the len of SortedList data struct
func (s *SortedList) SortedListSize() int {
	return s.List.Len()
}
