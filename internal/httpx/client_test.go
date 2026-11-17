package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientGetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := New(Config{Timeout: 5 * time.Second})
	code, body, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestClientGetNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := New(Config{Timeout: 5 * time.Second})
	code, _, err := c.Get(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", code)
	}
}

func TestMaxBody(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Write([]byte("0123456789"))
	}))
	defer srv.Close()

	if _, body, err := New(Config{MaxBody: 10}).Get(context.Background(), srv.URL); err != nil || len(body) != 10 {
		t.Fatalf("at cap: body %q, err %v", body, err)
	}
	hits = 0
	_, _, err := New(Config{MaxBody: 9}).Get(context.Background(), srv.URL)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("over cap: err = %v, want ErrBodyTooLarge", err)
	}
	if hits != 1 {
		t.Fatalf("over cap retried: %d requests", hits)
	}
}

func serve500(hits *int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(hits, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
}

func TestMaxRetriesOneSendsOneRequest(t *testing.T) {
	var hits int32
	srv := serve500(&hits)
	defer srv.Close()
	c := New(Config{MaxRetries: 1})
	code, _, err := c.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("want error")
	}
	if code != 500 {
		t.Fatalf("code = %d, want 500", code)
	}
	if hits != 1 {
		t.Fatalf("hits = %d, want 1", hits)
	}
}

func TestDefaultRetriesStillThree(t *testing.T) {
	var hits int32
	srv := serve500(&hits)
	defer srv.Close()
	c := New(Config{})
	code, _, err := c.Get(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("want error")
	}
	if code != 500 {
		t.Fatalf("code = %d, want 500", code)
	}
	if hits != 3 {
		t.Fatalf("hits = %d, want 3", hits)
	}
}

func TestLimiterAppliesPerAttempt(t *testing.T) {
	var hits int32
	srv := serve500(&hits)
	defer srv.Close()
	// 500ms limiter interval beats the 100ms backoff, so only a per-attempt
	// Wait makes 2 attempts take >= ~500ms.
	c := New(Config{Rate: 2, Burst: 1, MaxRetries: 2})
	start := time.Now()
	c.Get(context.Background(), srv.URL)
	if elapsed := time.Since(start); elapsed < 450*time.Millisecond {
		t.Fatalf("elapsed = %v, want >= 450ms", elapsed)
	}
	if n := atomic.LoadInt32(&hits); n != 2 {
		t.Fatalf("hits = %d, want 2", n)
	}
}
