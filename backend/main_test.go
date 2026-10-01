package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	h := handler("8081")
	for want := uint64(1); want <= 2; want++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/visits", nil))
		var got struct {
			Instance string `json:"instance"`
			Visits   uint64 `json:"visits"`
		}
		if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &got) != nil || got.Instance != "8081" || got.Visits != want {
			t.Fatalf("got status %d, body %s; want instance 8081, visits %d", w.Code, w.Body.String(), want)
		}
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("health status = %d, want 204", w.Code)
	}
}
