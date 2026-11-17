package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestDoSurfacesHeader pins that a 429's Retry-After reaches the
// caller. Without it a backoff has to guess how long the server wants us gone,
// and guessing short keeps the cooldown alive indefinitely.
func TestDoSurfacesHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "42")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := New(Config{Timeout: 5 * time.Second})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	code, _, retryAfter, err := c.Do("op", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != http.StatusTooManyRequests {
		t.Fatalf("code = %d, want 429", code)
	}
	if retryAfter != 42 {
		t.Fatalf("retryAfter = %d, want 42", retryAfter)
	}
}

// A response without the header reports 0, not an error: most responses carry none.
func TestDoAbsent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(Config{Timeout: 5 * time.Second})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	_, _, retryAfter, err := c.Do("op", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retryAfter != 0 {
		t.Fatalf("retryAfter = %d, want 0", retryAfter)
	}
}

// Do returns status and body alongside retryAfter.
func TestDoStillWorks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(Config{Timeout: 5 * time.Second})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	code, body, _, err := c.Do("op", req)
	if err != nil || code != http.StatusOK || string(body) != `{"ok":true}` {
		t.Fatalf("Do regressed: code=%d body=%q err=%v", code, body, err)
	}
}
