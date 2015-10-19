package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/abhi-bit/gouch"
)

var indexFileInfo *gouch.Gouch

type cache struct {
	channel chan string
	pos     int
	//cacheSize controls the cache size before read rows are written to the socket
	cacheSize int
}

var flusher http.Flusher
var c cache

func allDocumentsCallback(g *gouch.Gouch, docInfo *gouch.DocumentInfo, userContext interface{}, w io.Writer) error {
	row := "{\"id\":\"" + string(docInfo.ID) + "\",\"key\":" +
		string(docInfo.Key) + ",\"value\":" + string(docInfo.Value) + "}"
	userContext.(map[string]int)["count"]++

	c.channel <- string(row) + ",\n"
	c.pos++
	if c.pos == c.cacheSize {
		var buffer bytes.Buffer
		for i := 0; i < c.cacheSize; i++ {
			buffer.WriteString(<-c.channel)
		}
		c.pos = 0
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, buffer.String())
		flusher.Flush()
	}

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

	//Chunk in batches
	if limit > 20 {
		c.cacheSize = 20
	} else {
		c.cacheSize = limit
	}
	c.channel = make(chan string, c.cacheSize)

	now := time.Now()
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
	fmt.Fprintf(w, "Time elapsed: %v\n", time.Since(now))
}

func main() {

	indexFileInfo = &gouch.Gouch{}
	indexFileInfo.SetStatus(false)

	c = cache{pos: 0}

	http.HandleFunc("/query", runQuery)
	fmt.Println("Starting query prototype on port 9093")
	if err := http.ListenAndServe(":9093", nil); err != nil {
		log.Fatal(err)
	}
}
