package gouch

import (
	"fmt"
	"os"
)

type Gouch struct {
	file   *os.File
	pos    int64
	header *indexHeader
	ops    Ops
}

func Open(filename string, options int) (*Gouch, error) {
	return OpenEx(filename, options, NewBaseOps())
}

func OpenEx(filename string, options int, ops Ops) (*Gouch, error) {
	gouch := Gouch{
		ops: ops,
	}

	file, err := gouch.ops.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Errorf("Failed to open file: %v\n", filename)
		return nil, err
	}
	gouch.file = file

	fmt.Println("Trying to read from EOF")
	gouch.pos, err = gouch.ops.GotoEOF(gouch.file)
	if err != nil {
		fmt.Errorf("Failed while reading file from the end. file: %v\n", filename)
		return nil, err
	}
	fmt.Println("Gouch handler: %+v\n", gouch)

	if gouch.pos == 0 {
		fmt.Errorf("Empty file: %v\n", filename)
		return nil, err
	} else {
		fmt.Println("Trying to read last header")
		err = gouch.findLastHeader()
		if err != nil {
			return nil, err
		}
	}
	return &gouch, nil
}
