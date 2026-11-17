package kaize

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

const kaizeListHTML = `<!DOCTYPE html><html><body>
<div class="anime-list-element">
	<a class="name" href="/anime/cowboy-bebop">Cowboy Bebop</a>
</div>
<div class="anime-list-element">
	<a class="name" href="/anime/monster">Monster</a>
</div>
</body></html>`

func TestSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(kaizeListHTML))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	results, err := c.Search(context.Background(), "", types.SearchOpts{Page: 1})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Titles[0].Text != "Cowboy Bebop" {
		t.Fatalf("expected title Cowboy Bebop, got %s", results[0].Titles[0].Text)
	}
}

func TestGetNotImplemented(t *testing.T) {
	c := NewWithTestServer("http://dummy", Config{Timeout: 5 * time.Second})
	_, err := c.Get(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error for Get")
	}
}

func TestSearchForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	_, err := c.Search(context.Background(), "", types.SearchOpts{Page: 1})
	if !errors.Is(err, types.ErrRateLimited) {
		t.Fatalf("want ErrRateLimited (WAF block), got %v", err)
	}
	if types.StatusOf(err) != http.StatusForbidden {
		t.Fatalf("want status 403, got %d", types.StatusOf(err))
	}
}
