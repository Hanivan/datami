package hikka

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/linkeunid/datami/types"
)

func TestSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"pagination": {"total": 2, "pages": 1},
			"list": [
				{"slug": "cowboy-bebop", "mal_id": 1, "title_en": "Cowboy Bebop", "title_ja": "カウボーイビバップ", "title_ua": ""},
				{"slug": "monster", "mal_id": 2, "title_en": "Monster", "title_ja": "モンスター", "title_ua": ""}
			]
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	results, err := c.Search(context.Background(), "", types.SearchOpts{Limit: 100})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Titles[0].Text != "Cowboy Bebop" {
		t.Fatalf("expected title Cowboy Bebop, got %s", results[0].Titles[0].Text)
	}
	if len(results[0].ExternalIDs) != 1 || results[0].ExternalIDs[0].Source != string(types.SourceMAL) {
		t.Fatalf("expected external ID %s, got %v", types.SourceMAL, results[0].ExternalIDs)
	}
}

func TestGetNotImplemented(t *testing.T) {
	c := NewWithTestServer("http://dummy", Config{Timeout: 5 * time.Second})
	_, err := c.Get(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error for Get")
	}
}
