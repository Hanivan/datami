package provider_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hanivan/datami/provider/anilist"
	"github.com/Hanivan/datami/provider/mal"
)

type countLimiter struct{ n atomic.Int32 }

func (l *countLimiter) Wait(context.Context) error { l.n.Add(1); return nil }

func srv() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":1,"title":"x","data":{"Media":{"id":1,"title":{"romaji":"x"}}}}`))
	}))
}

func TestCustomLimiterIsUsed(t *testing.T) {
	s := srv()
	defer s.Close()
	ctx := context.Background()

	ml := &countLimiter{}
	m := mal.NewWithTestServer(s.URL, mal.Config{Limiter: ml})
	for i := 0; i < 3; i++ {
		m.Get(ctx, "1")
	}
	if got := ml.n.Load(); got != 3 {
		t.Fatalf("mal Wait calls = %d, want 3", got)
	}

	al := &countLimiter{}
	a := anilist.NewWithTestServer(s.URL, anilist.Config{Limiter: al})
	for i := 0; i < 3; i++ {
		a.Get(ctx, "1")
	}
	if got := al.n.Load(); got != 3 {
		t.Fatalf("anilist Wait calls = %d, want 3", got)
	}
}

// Nil Limiter must keep the built-in one: 3 requests at Rate 2/burst 1
// cannot all finish instantly, and must not panic.
func TestNilLimiterKeepsBuiltIn(t *testing.T) {
	s := srv()
	defer s.Close()
	m := mal.NewWithTestServer(s.URL, mal.Config{})
	start := time.Now()
	for i := 0; i < 3; i++ {
		m.Get(context.Background(), "1")
	}
	if d := time.Now().Sub(start); d < 500*1e6 {
		t.Fatalf("3 requests took %v; built-in limiter not applied", d)
	}
}
