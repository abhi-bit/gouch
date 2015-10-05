package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/abhi-bit/gouch"
)

func main() {
	order := func(i1, i2 *gouch.Item) int {
		var l int

		l1 := len(i1.Data)
		l2 := len(i2.Data)
		if l1 < l2 {
			l = l1
		} else {
			l = l2
		}
		return bytes.Compare(i1.Data[:l], i2.Data[:l])
	}

	// godebug
	_ = "breakpoint"

	list := gouch.SortedListCreate(order)
	bs1 := make([]byte, 4)
	binary.BigEndian.PutUint32(bs1, 1234567890)
	list.SortedListAdd(&gouch.Item{Data: bs1})

	bs2 := make([]byte, 4)
	binary.BigEndian.PutUint32(bs2, 2234567890)
	list.SortedListAdd(&gouch.Item{Data: bs2})

	element := list.SortedListGet(&gouch.Item{Data: bs2})
	if element != nil {
		fmt.Println("Found entry:", binary.BigEndian.Uint32(element.Data))
	}

	list.SortedListRemove(&gouch.Item{Data: bs2})

	bs3 := make([]byte, 4)
	binary.BigEndian.PutUint32(bs3, 4234567890)
	list.SortedListAdd(&gouch.Item{Data: bs3})

	fmt.Println("\nSorted list dump:")
	list.SortedListDump()

	bitmap := gouch.CreateBitmap()
	bitmap.SetBit(100)
	state := bitmap.GetBit(100)
	if state != true {
		fmt.Println("Issue! Bitmap entry missing")
	} else {
		fmt.Println("Bitmap entry found!")
	}
	fmt.Printf("%+v\n", bitmap.Dump())

	fmt.Println("\nFun starts here")
	gouch, err := gouch.Open("0c60b3073925d69702ce52efc90a9c4e.view.1", os.O_RDONLY)
	if err != nil {
		fmt.Errorf("Crashed while opening file\n")
	}
	fmt.Printf("Handler: %+v\n", gouch)
}
