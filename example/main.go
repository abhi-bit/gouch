package main

import (
	//"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/abhi-bit/gouch"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to a file")

func allDocumentsCallback(g *gouch.Gouch, docInfo *gouch.DocumentInfo, userContext interface{}, w io.Writer) error {
	//bytes, err := json.Marshal(docInfo)
	bytes := "{\"id\":\"" + string(docInfo.ID) + "\",\"key\":" + string(docInfo.Key) + ",\"value\":" + string(docInfo.Value) + "},"
	userContext.(map[string]int)["count"]++
	fmt.Println(bytes)
	return nil
}

func main() {
	// godebug
	_ = "breakpoint"
	//res := make(chan int)
	runtime.GOMAXPROCS(8)
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var w io.Writer
	//context := map[string]int{"count": 0}

	//g, err := gouch.Open("./index_with_1_entry", os.O_RDONLY)

	//100K records
	g, err := gouch.Open("./abhi_pymc_index", os.O_RDONLY)
	//g, err := gouch.Open("./composite_key_default", os.O_RDONLY)

	//1M records
	//g, err := gouch.Open("1M_pymc_index", os.O_RDONLY)

	if err != nil {
		fmt.Errorf("Crashed while opening file\n")
	}

	defer g.Close()
	//By-Id Btree
	//err = g.AllDocuments("", "", 100, allDocumentsCallback, context, w)
	//err = g.AllDocuments("", "", 100, allDocumentsCallback, context, w)

	//Map-reduce Btree
	/*for i := 0; i < 100; i++ {
		go func(i int) {
			context := map[string]int{"count": 0}
			err = g.AllDocsMapReduce("", "", allDocumentsCallback, context, w, 10)
			res <- i
		}(i)
	}

	for i := 0; i < 100; i++ {
		select {
		case resi := <-res:
			fmt.Println(resi)
		}
	}*/

	f, _ := os.Create("mprof")
	pprof.WriteHeapProfile(f)
	f.Close()

	context := map[string]int{"count": 0}
	err = g.AllDocsMapReduce("", "", allDocumentsCallback, context, w, 10)
}
