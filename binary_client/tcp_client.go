package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
	//"strconv"
	"sync"
	"sync/atomic"
	"time"

	//"github.com/abhi-bit/gouch"
)

var goroutines = flag.Int("t", 100, "Worker threads to spawn")
var limit = flag.Int("l", 10, "Limit in view query call, default 10")
var numRequests = flag.Int("r", 100, "Number of requests per worker thread")
var size = flag.Int("s", 100, "Size of string")
var wg sync.WaitGroup

func randString() string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, *size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())
	wg.Add(*goroutines)

	var totalTime uint64
	var totalDataRead uint64

	text := randString()

	now := time.Now()
	for i := 0; i < *goroutines; i++ {
		go func() {
			conn, err := net.Dial("tcp", "127.0.0.1:9091")
			defer conn.Close()
			if err != nil {
				log.Println(err)
			}

			dataReadSize := 0
			for j := 0; j < *numRequests; j++ {
				//text := "GET limit " + strconv.Itoa(*limit)
				fmt.Fprintf(conn, text+"\n")
				message, _ := bufio.NewReader(conn).ReadString('\n')
				dataReadSize += len(message)
			}
			atomic.AddUint64(&totalDataRead, uint64(dataReadSize))
			runtime.Gosched()
			wg.Done()

		}()
	}
	wg.Wait()

	diff := time.Since(now)
	atomic.AddUint64(&totalTime, uint64(diff))
	//in KB
	dataRead := float64(totalDataRead) / (1000)
	//in ms
	timeTaken := float64(totalTime) / (1000 * 1000)

	fmt.Printf("total time: %f ms\n", timeTaken)
	fmt.Printf("total data read: %f KB\n", dataRead)
	fmt.Printf("Requests/sec: %f\n", float64(*goroutines**numRequests)/(timeTaken)*1000)
	fmt.Printf("throughput: %f MB/s\n", (dataRead / timeTaken))
}
