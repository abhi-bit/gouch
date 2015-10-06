package main

import (
	"bytes"
	//"encoding/binary"
	"encoding/json"
	//	"flag"
	"fmt"
	"os"

	"github.com/abhi-bit/gouch"
)

//var filename = flag.String("file", "", "view index file to read")

func allDocumentsCallback(g *gouch.Gouch, docInfo *gouch.DocumentInfo, userContext interface{}) error {
	bytes, err := json.MarshalIndent(docInfo, "", "  ")
	if err != nil {
		fmt.Println(err)
	} else {
		userContext.(map[string]int)["count"]++
		fmt.Println(string(bytes))
	}
	return nil
}

func main() {
	//	flag.Parse()

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

	_ = gouch.SortedListCreate(order)
	/*list := gouch.SortedListCreate(order)
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
	fmt.Printf("%+v\n", bitmap.Dump())*/

	//gouch, err := gouch.Open(*filename, os.O_RDONLY)

	context := map[string]int{"count": 0}

	//g, err := gouch.Open("vbucket", os.O_RDONLY)
	//g, err := gouch.Open("index", os.O_RDONLY)
	//g, err := gouch.Open("index_with_1_entry", os.O_RDONLY)
	//g, err := gouch.Open("index_25K_items_4_node", os.O_RDONLY)
	g, err := gouch.Open("index_with_10K_entries", os.O_RDONLY)
	//g, err := gouch.Open("beer_sample", os.O_RDONLY)
	if err != nil {
		fmt.Errorf("Crashed while opening file\n")
	}
	fmt.Printf("Handler: %+v\n", g)

	err = g.AllDocuments("", "", allDocumentsCallback, context)

}
