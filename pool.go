package gouch

import (
	"sync"
)

//SnappyDecodeBufLen used for sync pooling decoding byte chunks via snappy
const SnappyDecodeBufLen int = 21000

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

var eightByte = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 8)
	},
}

var snappyDecodeChunk = &sync.Pool{
	New: func() interface{} {
		return make([]byte, SnappyDecodeBufLen)
	},
}
