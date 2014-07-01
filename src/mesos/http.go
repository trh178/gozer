package mesos

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	port = 8888
)

var (
//	port       = flag.Int("port", 4242, "Port to listen on for HTTP endpoint")
)

func init() {
	http.HandleFunc("/", rootHandler)
	httpWaitGroup.Add(1)
	go startServing()
}

func startServing() {
	log.Printf("listening on port %d", port)
	httpWaitGroup.Done()
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("failed to start listening on port %d", port)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("received request with unexpected method. want %q, got %q: %+v", "POST", r.Method, r)
		return
	}

	pathElements := strings.Split(r.URL.Path, "/")

	if pathElements[1] != frameworkName {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("unexpected path. want %q, got %q", frameworkName, pathElements[1])))
		log.Printf("received request with unexpected path. want %q, got %q: %+v", frameworkName, pathElements[1], r)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: failed to read body from request %+v: %+v", r, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	event, err := bytesToEvent(pathElements[2], body)
	if err != nil {
		log.Printf("ERROR: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	events <- event

	w.WriteHeader(http.StatusOK)
}
