package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestServerPool(t *testing.T) {
	pool := &ServerPool{}
	for _, host := range []string{"one.example", "two.example"} {
		u := &url.URL{Scheme: "http", Host: host}
		proxy := &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) { r.SetURL(u) },
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(r.URL.Host)),
				}, nil
			}),
		}
		pool.backends = append(pool.backends, &Backend{URL: u, Alive: true, ReverseProxy: proxy})
	}

	check := func(wantCode int, wantBody string) {
		t.Helper()
		w := httptest.NewRecorder()
		pool.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://lb.example/path", nil))
		if w.Code != wantCode || (wantBody != "" && strings.TrimSpace(w.Body.String()) != wantBody) {
			t.Fatalf("got %d %q, want %d %q", w.Code, w.Body.String(), wantCode, wantBody)
		}
	}

	check(http.StatusOK, "one.example")
	check(http.StatusOK, "two.example")
	check(http.StatusOK, "one.example")
	pool.backends[0].SetAlive(false)
	check(http.StatusOK, "two.example")
	pool.backends[1].SetAlive(false)
	check(http.StatusServiceUnavailable, "Service not available")
}
