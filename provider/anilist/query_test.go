package anilist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

// TestQueriesSubSelectObjectFields pins the GraphQL contract: rankings
// ([MediaRank]) and stats (MediaStats) are object types, so AniList rejects
// every query that requests them bare with "must have a sub selection" and a
// 400 — which takes the whole request down, not just those two fields.
func TestQueriesSubSelectObjectFields(t *testing.T) {
	var captured []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		captured = append(captured, string(buf))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"Media":{"id":1,"title":{"romaji":"x"}},"Page":{"Media":[]}}}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	ctx := context.Background()
	c.Get(ctx, "1")
	c.Search(ctx, "x", types.SearchOpts{})
	c.GetByMalID(ctx, "1")
	c.GetManga(ctx, "1")
	c.SearchManga(ctx, "x", types.SearchOpts{})
	c.GetMangaByMalID(ctx, "1")

	if len(captured) == 0 {
		t.Fatal("no queries captured")
	}
	for i, q := range captured {
		if strings.Contains(q, "rankings stats") || strings.Contains(q, "rankings }") {
			t.Errorf("query %d requests rankings/stats without a sub selection: AniList answers 400", i)
		}
		if strings.Contains(q, "rankings") && !strings.Contains(q, "rankings {") {
			t.Errorf("query %d: rankings needs a sub selection", i)
		}
		if strings.Contains(q, "stats") && !strings.Contains(q, "stats {") {
			t.Errorf("query %d: stats needs a sub selection", i)
		}
	}
}

func TestConfigUserAgentReachesRequest(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		w.Write([]byte(`{"data":{"Media":{"id":1}}}`))
	}))
	defer srv.Close()
	c := NewWithTestServer(srv.URL, Config{UserAgent: "Newda/1.0"})
	_, _ = c.Get(context.Background(), "1")
	if got != "Newda/1.0" {
		t.Fatalf("User-Agent = %q", got)
	}
}
