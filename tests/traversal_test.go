package gouch_test

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/abhi-bit/gouch"
)

func BenchmarkTraversal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g, err := gouch.Open("../example/1M_pymc_index", os.O_RDONLY)

		context := map[string]int{"count": 0}
		var w io.Writer
		err = g.AllDocsMapReduce("", "", gouch.DefaultDocumentCallback, context, w, 10)
		if err != nil {
			log.Fatal(err)
		}
	}
}
