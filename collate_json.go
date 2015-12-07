package gouch

/*
#include "kway_merge/collate_json.h"
#include "kway_merge/min_heap.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func collateJSON(a, b []byte) int {
	c1 := C.CString(string(a))
	s1 := C.size_t(len(a))
	sb1 := C.sized_buf{buf: c1, size: s1}

	c2 := C.CString(string(a))
	s2 := C.size_t(len(a))
	sb2 := C.sized_buf{buf: c2, size: s2}

	fmt.Printf("sb1: %#v sb2: %#v\n", sb1, sb2)

	ret := int(C.CollateJSON(&sb1, &sb2, 0))
	C.free(unsafe.Pointer(c1))
	C.free(unsafe.Pointer(c2))

	return ret
}
