package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

func handler(instance string) http.Handler {
	var visits atomic.Uint64
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /api/visits", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Instance string `json:"instance"`
			Visits   uint64 `json:"visits"`
		}{instance, visits.Add(1)})
	})

	return mux
}

func main() {
	port := "8080"

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	log.Printf("backend %s listening", port)
	log.Fatal(http.ListenAndServe(":"+port, handler(port)))
}
