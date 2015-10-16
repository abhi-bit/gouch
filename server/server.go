package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/abhi-bit/gouch"
)

var limit int

func allDocumentsCallback(g *gouch.Gouch, docInfo *gouch.DocumentInfo, userContext interface{}, w io.Writer) error {
	bytes := "{\"id\":\"" + string(docInfo.ID) + "\",\"key\":" +
		string(docInfo.Key) + ",\"value\":" + string(docInfo.Value) + "},"
	userContext.(map[string]int)["count"]++
	fmt.Fprintf(w, string(bytes)+",\n")
	return nil
}

func runQuery(w http.ResponseWriter, r *http.Request) {
	_, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")

	startKey := ""
	endKey := ""
	var limit int

	params := r.URL.Query()
	if limitParam, ok := params["limit"]; ok && len(limitParam) > 0 {
		limit, _ = strconv.Atoi(limitParam[0])
	}
	if start, ok := params["start"]; ok && len(start) > 0 {
		startKey = start[0]
	}
	if end, ok := params["end"]; ok && len(end) > 0 {
		endKey = end[0]
	}

	//fmt.Printf("startKey: %s endKey: %s limit: %d\n", startKey, endKey, limit)

	context := map[string]int{"count": 0}

	now := time.Now()
	g, _ := gouch.Open("/tmp/1M_pymc_index", os.O_RDONLY)
	//g, _ := gouch.Open("/Users/asingh/repo/go/src/github.com/abhi-bit/gouch/example/1M_pymc_index", os.O_RDONLY)
	err := g.AllDocsMapReduce(startKey, endKey, allDocumentsCallback, context, w, limit)
	if err != nil {
		fmt.Printf("Failed tree traversal\n")
	}
	g.Close()
	fmt.Fprintf(w, "Time elapsed: %v\n", time.Since(now))

}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	http.HandleFunc("/query", runQuery)
	fmt.Println("Starting query prototype on port 8093")
	if err := http.ListenAndServe(":9093", nil); err != nil {
		log.Fatal(err)
	}
}
