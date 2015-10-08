package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/abhi-bit/gouch"
)

func allDocumentsCallback(g *gouch.Gouch, docInfo *gouch.DocumentInfo, userContext interface{}, w io.Writer) error {
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
	// godebug
	_ = "breakpoint"

	var w io.Writer
	context := map[string]int{"count": 0}

	//100K records
	g, err := gouch.Open("pymc0_index", os.O_RDONLY)

	//1M records
	//g, err := gouch.Open("1M_pymc_index", os.O_RDONLY)
	if err != nil {
		fmt.Errorf("Crashed while opening file\n")
	}

	//By-Id Btree
	//err = g.AllDocuments("", "", allDocumentsCallback, context, w)

	//Map-reduce Btree
	err = g.AllDocsMapReduce("", "", allDocumentsCallback, context, w)

}
