package gouch

import (
	"sync"
)

//4 byte sync.Pool
var s *slicePool

//2 byte sync.Pool
var p *slicePool

type byteSliceFactory func() []byte

type genericPool struct {
	pool *sync.Pool
	//allocs records the number of allocations
	allocs int64
}

func newGenericPool(newfactory func() interface{}) *genericPool {
	g := &genericPool{
		allocs: 0,
	}
	g.pool = &sync.Pool{
		New: func() interface{} {
			g.allocs++
			return newfactory()
		},
	}
	return g
}

type slicePool struct {
	gp *genericPool
}

func newSlicePool(b byteSliceFactory) *slicePool {
	return &slicePool{
		gp: newGenericPool(func() interface{} { return b() }),
	}
}

func (s *slicePool) getBytes() []byte {
	b := s.gp.pool.Get().([]byte)
	return b[:cap(b)]
}

func (s *slicePool) putBytes(b []byte) {
	s.gp.pool.Put(b)
}
