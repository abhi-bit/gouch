package gouch

import (
	"sync"
)

var twoByte = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 2)
	},
}
var fourByte = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 4)
	},
}

var documentInfoPool = &sync.Pool{
	New: func() interface{} {
		return DocumentInfo{}
	},
}
