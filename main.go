package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync/atomic"
)

type Backend struct {
	URL          *url.URL
	ReverseProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func (s *ServerPool) GetNextPeer() *Backend {
	if len(s.backends) == 0 {
		return nil
	}

	next := (atomic.AddUint64(&s.current, 1) - 1) % uint64(len(s.backends))
	return s.backends[next]
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := s.GetNextPeer()
	if peer == nil {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}
	peer.ReverseProxy.ServeHTTP(w, r)
}

func main() {
	backendURLs := os.Args[1:]
	if len(backendURLs) == 0 {
		backendURLs = []string{"http://localhost:8080"}
	}

	pool := &ServerPool{}
	for _, address := range backendURLs {
		u, err := url.Parse(address)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
			log.Fatalf("invalid backend URL: %q", address)
		}
		proxy := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)
			r.SetXForwarded()
		}}
		pool.backends = append(pool.backends, &Backend{URL: u, ReverseProxy: proxy})
	}

	log.Printf("listening on :8000 with %d backend(s)", len(pool.backends))
	log.Fatal(http.ListenAndServe(":8000", pool))
}
