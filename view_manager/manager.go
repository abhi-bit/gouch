package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

//QueryTimeout set to 2 seconds
const QueryTimeout int = 2

var (
	address string
	hosts   string
	nodes   = make(map[string]nodeStatus)
	port    int
)

var tr = &http.Transport{
	MaxIdleConnsPerHost: 5000,
}

var client = &http.Client{Transport: tr}

func init() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&address, "address", "", "Address to listen on, Default is to all")
	flag.IntVar(&port, "port", 9091, "Port to listen on. Default is 8091")
	flag.StringVar(&hosts, "host", "localhost:9093", "nodes to manage")
	flag.Parse()

}

type nodeStatus struct {
	status  bool
	retries int
}

func getNodes(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(200)

	var onlineNodes []string
	for node := range nodes {
		onlineNodes = append(onlineNodes, node)
	}

	oNodes, _ := json.Marshal(onlineNodes)
	fmt.Fprintf(w, fmt.Sprintf("{\"nodes\":%s}", oNodes))
}

func runQuery(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	query := r.URL.RawQuery

	resc, errc := make(chan string), make(chan error)

	var wg sync.WaitGroup

	for node := range nodes {
		wg.Add(1)
		go func(node string) {
			url := node + path + "?" + query
			data, err := fetch(url)
			if err != nil {
				errc <- err
				return
			}
			resc <- data
			wg.Done()
		}(node)
	}

	for _ = range nodes {
		select {
		case res := <-resc:
			fmt.Fprintf(w, res)
		case err := <-errc:
			fmt.Fprintf(w, err.Error())
		}
	}

	wg.Wait()

	defer close(resc)
	defer close(errc)
}

func fetch(url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Sub-query call failed against %s: %+v\n", url, err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response from %s: %+v\n", url, err)
		return "", err
	}
	return string(body), nil
}
func main() {

	log.Printf("listening on %s:%d\n", address, port)

	for _, host := range strings.Split(hosts, ",") {
		serverURL := "http://" + host

		resp, err := http.Get(serverURL)

		if err != nil {
			log.Println(err)
			nodes[serverURL] = nodeStatus{status: false, retries: 0}
			break
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if string(body) == "1" {
			nodes[serverURL] = nodeStatus{status: true, retries: 0}
		} else {
			nodes[serverURL] = nodeStatus{status: false, retries: 0}
		}
	}

	fmt.Printf("%#v\n", nodes)

	//Polling nodes, needs cleanup
	go func() {
		for {
			for node := range nodes {
				resp, err := http.Get(node)
				if err != nil {
					log.Println(err)
					retryCount := nodes[node].retries + 1

					if retryCount <= 3 {
						nodes[node] = nodeStatus{status: false, retries: retryCount}
					} else {
						delete(nodes, node)
					}

					break
				}

				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if string(body) == "1" {
					nodes[node] = nodeStatus{status: true, retries: 0}
				} else {
					retryCount := nodes[node].retries + 1
					if retryCount <= 3 {
						nodes[node] = nodeStatus{status: false, retries: retryCount}
					} else {
						delete(nodes, node)
					}
				}
			}
			//fmt.Printf("%#v\n", nodes)
			time.Sleep(time.Second)
		}
	}()

	http.HandleFunc("/nodes", getNodes)
	http.HandleFunc("/query", runQuery)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", address, port), nil)
	if err != nil {
		log.Fatalf("Failed to start view manager: %v", err)
	}
}
