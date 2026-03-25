package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
)

var (
	count int
	mu    sync.Mutex
	port  = 5050
)

func tallyHandle(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	count++
	curr := count
	mu.Unlock()

	slog.Info("Tally incremented",
		"curr_count", curr,
		"remote_addr", r.RemoteAddr,
	)

	fmt.Fprintf(w, "%d", curr)
}

func currHandle(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	curr := count
	mu.Unlock()

	fmt.Fprintf(w, "%d", curr)
}

func main() {
	http.HandleFunc("/tally", tallyHandle)
	http.HandleFunc("/curr", currHandle)
	slog.Info("Server-",
		"port", port,
		"status", "ready",
	)

	err := (http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	if err != nil {
		log.Fatal(err)
	}
}
