package gouch

/*
#include "./kway_merge/collate_json.h"
#include "./kway_merge/min_heap.h"
*/
import "C"

func collateJSON(a, b []byte) int {
	b1 := C.CString(string(a))
	b2 := C.CString(string(b))

	s1 := C.int(len(a))
	s2 := C.int(len(b))

	return int(C.collate_JSON(b1, b2, s1, s2))
}
