package otakotaku

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

func TestGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"id_anime": 1,
				"judul_anime": "Cowboy Bebop",
				"mal_id_anime": 1,
				"anidb_id_anime": 1,
				"ann_id_anime": 1,
				"ap_id_anime": 1
			}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	m, err := c.Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if m.Titles[0].Text != "Cowboy Bebop" {
		t.Fatalf("expected title Cowboy Bebop, got %s", m.Titles[0].Text)
	}
	if len(m.ExternalIDs) != 4 {
		t.Fatalf("expected 4 external IDs, got %d", len(m.ExternalIDs))
	}
}

func TestSearchNotImplemented(t *testing.T) {
	c := NewWithTestServer("http://dummy", Config{Timeout: 5 * time.Second})
	_, err := c.Search(context.Background(), "test", types.SearchOpts{})
	if err == nil {
		t.Fatal("expected error for Search")
	}
}

func TestGetByMalIDNotImplemented(t *testing.T) {
	c := NewWithTestServer("http://unused", Config{})
	if _, err := c.GetByMalID(context.Background(), "1"); !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("want ErrNotImplemented, got %v", err)
	}
}
