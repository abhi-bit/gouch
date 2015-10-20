package gouch

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

//Cache used to cache some rows from view query calls
type Cache struct {
	channel   chan string
	pos       int
	cacheSize int
}

//CreateCache inits Cache
func CreateCache(pos int) *Cache {
	c := Cache{pos: 0}
	return &c
}

//SetCacheSize sets the buffer size of Cache
func (c *Cache) SetCacheSize(cacheSize int) {
	c.cacheSize = cacheSize
}

//SetCacheChannelBuffer creates buffered channels
//Typically chan size should be 5 times of cacheSize
func (c *Cache) SetCacheChannelBuffer(bufSize int) {
	c.channel = make(chan string, bufSize)
}

//CacheStore stores rows into the chan
func (c *Cache) CacheStore(value string, w io.Writer) {
	c.channel <- value
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
}
