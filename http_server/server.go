package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"

	"github.com/abhi-bit/gouch"
)

var indexFileInfo *gouch.Gouch
var port int
var flusher http.Flusher

func init() {
	flag.IntVar(&port, "port", 9093, "Port to listen. Default is 9093")
	flag.Parse()
}

func healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprint(w, 1)
}

func allDocumentsCallback(g *gouch.Gouch, docInfo *gouch.DocumentInfo, userContext interface{}, w io.Writer) error {
	row := "{\"id\":\"" + string(docInfo.ID) + "\",\"key\":" +
		string(docInfo.Key) + ",\"value\":" + string(docInfo.Value) + "}"
	userContext.(map[string]int)["count"]++

	flusher, _ := w.(http.Flusher)
	fmt.Fprintf(w, string(row)+"\n")
	flusher.Flush()

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

	context := map[string]int{"count": 0}

	//now := time.Now()
	var g *gouch.Gouch
	if indexFileInfo.GetFDStatus() == false {
		g, _ = gouch.Open("/tmp/1M_pymc_index", os.O_RDONLY)
		indexFileInfo = g.DeepCopy()
		indexFileInfo.SetStatus(true)
	} else {
		g = indexFileInfo.DeepCopy()
	}

	err := indexFileInfo.AllDocsMapReduce(startKey, endKey, allDocumentsCallback, context, w, limit)
	if err != nil {
		fmt.Printf("Failed tree traversal\n")
	}
	//Not closing FD so that we could reuse it
	//indexFileInfo.Close()
	//fmt.Fprintf(w, "Time elapsed: %v\n", time.Since(now))
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	indexFileInfo = &gouch.Gouch{}
	indexFileInfo.SetStatus(false)

	http.HandleFunc("/", healthCheck)
	http.HandleFunc("/query", runQuery)
	fmt.Printf("Starting query prototype on port %d\n", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal(err)
	}
}
