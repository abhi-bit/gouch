package main

import (
	"bufio"
	"log"
	"net"
	"runtime"
	"strings"

	"github.com/goinggo/workpool"
)

type work struct {
	conn net.Conn
	wp   *workpool.WorkPool
}

func (w *work) DoWork(workRoutine int) {
	m, _ := bufio.NewReader(w.conn).ReadString('\n')

	for len(m) != 0 {
		message := strings.ToUpper(m)
		w.conn.Write([]byte(message))
		m, _ = bufio.NewReader(w.conn).ReadString('\n')
	}

	w.conn.Close()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	workPool := workpool.New(runtime.NumCPU()*3, 1000)

	ln, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		w := work{
			conn: c,
			wp:   workPool,
		}

		if err := workPool.PostWork("routine", &w); err != nil {
			log.Println(err)
		}
	}
}
