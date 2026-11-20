package anilist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

type reqLog struct {
	queries []string
	vars    []map[string]any
}

func nestedServer(t *testing.T, log *reqLog, respond func(n int) string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var p struct {
			Query string         `json:"query"`
			Vars  map[string]any `json:"variables"`
		}
		json.Unmarshal(b, &p)
		log.queries = append(log.queries, p.Query)
		log.vars = append(log.vars, p.Vars)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, respond(len(log.queries)))
	}))
	t.Cleanup(srv.Close)
	return NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
}

func cEdge(id int) string {
	return fmt.Sprintf(`{"role":"MAIN","node":{"id":%d,"name":{"full":"c%d"}}}`, id, id)
}
func sEdge(id int) string {
	return fmt.Sprintf(`{"role":"Director","node":{"id":%d,"name":{"full":"s%d"}}}`, id, id)
}
func conn(next bool, edges ...string) string {
	return fmt.Sprintf(`{"pageInfo":{"hasNextPage":%t},"edges":[%s]}`, next, strings.Join(edges, ","))
}

func TestGetByMalID_EmbedsPages(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string {
		return `{"data":{"Media":{"id":7,"title":{"romaji":"x"},
			"c1":` + conn(true, cEdge(1), cEdge(2)) + `,"c2":` + conn(false, cEdge(3)) + `,
			"s1":` + conn(true, sEdge(1)) + `,"s2":` + conn(true, sEdge(2)) + `,"s3":` + conn(false, sEdge(3)) + `}}}`
	})
	m, err := c.GetByMalID(context.Background(), "1", types.MediaOpts{CharacterPages: 2, StaffPages: 3})
	if err != nil {
		t.Fatal(err)
	}
	q := log.queries[0]
	for _, want := range []string{"c1: characters(page: 1, perPage: 25, sort: ROLE)", "c2: characters(page: 2", "s1: staff(page: 1, perPage: 25, sort: RELEVANCE)", "s3: staff(page: 3"} {
		if !strings.Contains(q, want) {
			t.Errorf("query missing %q", want)
		}
	}
	if strings.Contains(q, "c3:") || strings.Contains(q, "s4:") {
		t.Error("query has extra aliases")
	}
	if len(m.Characters) != 3 || len(m.Staff) != 3 {
		t.Fatalf("chars=%d staff=%d", len(m.Characters), len(m.Staff))
	}
	for i, want := range []string{"1", "2", "3"} {
		if m.Characters[i].ID != want || m.Staff[i].ID != want {
			t.Errorf("order at %d: char %q staff %q, want %q", i, m.Characters[i].ID, m.Staff[i].ID, want)
		}
	}
	if len(log.queries) != 1 {
		t.Errorf("requests=%d", len(log.queries))
	}
}

func TestGetByMalID_FollowUpWhenMore(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(n int) string {
		if n == 1 {
			return `{"data":{"Media":{"id":7,"title":{"romaji":"x"},"c1":` + conn(true, cEdge(1)) + `,"c2":` + conn(true, cEdge(2)) + `}}}`
		}
		return `{"data":{"Media":{"characters":` + conn(false, cEdge(3)) + `}}}`
	})
	m, err := c.GetByMalID(context.Background(), "1", types.MediaOpts{CharacterPages: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(log.queries) != 2 {
		t.Fatalf("requests=%d", len(log.queries))
	}
	if !strings.Contains(log.queries[1], "characters(page: $page") {
		t.Errorf("follow-up not standalone: %s", log.queries[1])
	}
	if log.vars[1]["page"] != float64(3) || log.vars[1]["id"] != float64(7) {
		t.Errorf("vars=%v", log.vars[1])
	}
	if len(m.Characters) != 3 {
		t.Fatalf("chars=%d", len(m.Characters))
	}
	for i, want := range []string{"1", "2", "3"} {
		if m.Characters[i].ID != want {
			t.Errorf("order at %d: %q, want %q", i, m.Characters[i].ID, want)
		}
	}
}

func TestGetByMalID_EmbeddedPagesNoNext(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string {
		return `{"data":{"Media":{"id":7,"title":{"romaji":"x"},"c1":` + conn(false) + `}}}`
	})
	m, err := c.GetByMalID(context.Background(), "1", types.MediaOpts{CharacterPages: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Characters) != 0 || len(log.queries) != 1 {
		t.Errorf("chars=%d requests=%d", len(m.Characters), len(log.queries))
	}
}

func TestGetByMalID_NoOptsUnchanged(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return `{"data":{"Media":{"id":7,"title":{"romaji":"x"}}}}` })
	m, err := c.GetByMalID(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(log.queries[0], "characters(") || strings.Contains(log.queries[0], "staff(") || m.Characters != nil {
		t.Error("embedded without opts")
	}
}

func TestGetByMalID_NotFoundWithOpts(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return `{"data":{"Media":null}}` })
	_, err := c.GetByMalID(context.Background(), "1", types.MediaOpts{CharacterPages: 2})
	if !errors.Is(err, types.ErrNotFound) || len(log.queries) != 1 {
		t.Errorf("err=%v requests=%d", err, len(log.queries))
	}
}

func TestGetMangaByMalID_NoStaffAlias(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return `{"data":{"Media":{"id":7,"title":{"romaji":"x"}}}}` })
	if _, err := c.GetMangaByMalID(context.Background(), "1", types.MediaOpts{CharacterPages: 2}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(log.queries[0], "s1:") || !strings.Contains(log.queries[0], "c1:") {
		t.Error("alias selection wrong")
	}
}

func TestGetCharacters_PageOpts(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string {
		return `{"data":{"Media":{"characters":` + conn(true, cEdge(1)) + `}}}`
	})
	if _, err := c.GetCharacters(context.Background(), "7", types.PageOpts{StartPage: 3, MaxPages: 1}); err != nil {
		t.Fatal(err)
	}
	if len(log.queries) != 1 || log.vars[0]["page"] != float64(3) {
		t.Errorf("requests=%d vars=%v", len(log.queries), log.vars)
	}
}

func pagesRequested(log *reqLog) []float64 {
	var out []float64
	for _, v := range log.vars {
		out = append(out, v["page"].(float64))
	}
	return out
}

func TestGetCharacters_PageOptsClamped(t *testing.T) {
	always := func(int) string { return `{"data":{"Media":{"characters":` + conn(true, cEdge(1)) + `}}}` }

	var log reqLog
	c := nestedServer(t, &log, always)
	if _, err := c.GetCharacters(context.Background(), "7", types.PageOpts{StartPage: 9, MaxPages: 5}); err != nil {
		t.Fatal(err)
	}
	if got := pagesRequested(&log); len(got) != 2 || got[0] != 9 || got[1] != 10 {
		t.Errorf("pages=%v, want [9 10]", got)
	}

	var log2 reqLog
	c = nestedServer(t, &log2, always)
	if _, err := c.GetCharacters(context.Background(), "7", types.PageOpts{MaxPages: 50}); err != nil {
		t.Fatal(err)
	}
	if len(log2.queries) != 10 {
		t.Errorf("requests=%d, want 10", len(log2.queries))
	}

	var log3 reqLog
	c = nestedServer(t, &log3, always)
	if _, err := c.GetCharacters(context.Background(), "7", types.PageOpts{StartPage: 11}); err != nil {
		t.Fatal(err)
	}
	if len(log3.queries) != 0 {
		t.Errorf("requests=%d, want 0", len(log3.queries))
	}
}

// embedPages trims the trailing " } }" of a media query; a query that stops
// ending that way would silently embed into the wrong place.
func TestMediaQueriesEndWithMediaClose(t *testing.T) {
	var log reqLog
	c := nestedServer(t, &log, func(int) string { return `{"data":{"Media":{"id":1,"title":{"romaji":"x"}}}}` })
	ctx := context.Background()
	c.Get(ctx, "1")
	c.GetManga(ctx, "1")
	c.GetByMalID(ctx, "1")
	c.GetMangaByMalID(ctx, "1")
	if len(log.queries) != 4 {
		t.Fatalf("queries=%d", len(log.queries))
	}
	for i, q := range log.queries {
		if !strings.HasSuffix(q, " } }") {
			t.Errorf("query %d does not end with \" } }\"", i)
		}
	}
}
