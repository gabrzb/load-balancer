package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

func (s *ServerPool) GetNextPeer() *Backend {
	if len(s.backends) == 0 {
		return nil
	}

	next := int((atomic.AddUint64(&s.current, 1) - 1) % uint64(len(s.backends)))
	for i := 0; i < len(s.backends); i++ {
		b := s.backends[(next+i)%len(s.backends)]
		if b.IsAlive() {
			return b
		}
	}

	return nil
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		log.Printf("%s alive: %t", b.URL, alive)
	}
}

func isBackendAlive(u *url.URL) bool {
	port := u.Port()

	if port == "" {
		port = "80"
		if u.Scheme == "https" {
			port = "443"
		}
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(u.Hostname(), port), 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
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
		backend := &Backend{URL: u}
		proxy := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)
			r.SetXForwarded()
		}}
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("backend %s: %v", u, err)
			backend.SetAlive(false)
			http.Error(w, "Service not available", http.StatusBadGateway)
		}
		backend.ReverseProxy = proxy
		pool.backends = append(pool.backends, backend)
	}

	pool.HealthCheck()

	go func() {
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			pool.HealthCheck()
		}
	}()

	log.Printf("listening on :8000 with %d backend(s)", len(pool.backends))
	log.Fatal(http.ListenAndServe(":8000", pool))
}
