package main

import (
	"fmt"
	"github.com/abhi-bit/gouch"
	"os"
)

func main() {

	g, err := gouch.Open("./index.1", os.O_RDONLY)
	if err != nil {
		fmt.Println("Index file open errored")
	} else {
		fmt.Println("Index file open success")
	}
	fmt.Printf("Header count: %d\n", g.GetHeaderCount())
	g.Close()
}
