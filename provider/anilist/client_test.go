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

func TestGetMedia(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"Media": {
					"id": 1,
					"title": {"romaji": "Cowboy Bebop", "english": "Cowboy Bebop", "native": "カウボーイビバップ"},
					"description": "Space bounty hunters.",
					"startDate": {"year": 1998, "month": 4, "day": 3},
					"endDate": {"year": 1998, "month": 6, "day": 26},
					"status": "FINISHED",
					"format": "TV",
					"episodes": 26,
					"duration": 24,
					"averageScore": 78,
					"meanScore": 78,
					"popularity": 15000,
					"genres": ["Action", "Sci-Fi"],
					"studios": {"nodes": [{"id": 14, "name": "Sunrise"}]},
					"coverImage": {"medium": "https://example.com/cb.jpg", "large": "https://example.com/cb_large.jpg"},
					"relations": {"edges": [{"relationType": "SEQUEL", "node": {"id": 2, "title": {"romaji": "Cowboy Bebop: The Movie"}}}]},
					"externalLinks": [],
					"synonyms": ["Bebop"]
				}
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
	if m.EpisodeCount != 26 {
		t.Fatalf("expected 26 episodes, got %d", m.EpisodeCount)
	}
	if m.Status != "finished" {
		t.Fatalf("expected status finished, got %s", m.Status)
	}
	if m.Score == nil || m.Score.Mean != 78 {
		t.Fatal("score mean mismatch")
	}
	if len(m.Genres) != 2 || m.Genres[0] != "Action" {
		t.Fatalf("genres mismatch: %v", m.Genres)
	}
	if len(m.Studios) != 1 || m.Studios[0].Name != "Sunrise" {
		t.Fatalf("studios mismatch: %v", m.Studios)
	}
	if m.StartDate == nil || *m.StartDate.Year != 1998 {
		t.Fatal("start date year mismatch")
	}
}

func TestGetNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	_, err := c.Get(context.Background(), "999999")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestGetManga(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"Media": {
					"id": 1,
					"title": {"romaji": "Monster", "english": "Monster", "native": "モンスター"},
					"description": "A thriller.",
					"startDate": {"year": 1994, "month": 12, "day": 5},
					"endDate": {"year": 2001, "month": 12, "day": 20},
					"status": "FINISHED",
					"format": "MANGA",
					"chapters": 162,
					"volumes": 18,
					"averageScore": 91,
					"meanScore": 91,
					"popularity": 5000,
					"genres": ["Mystery", "Psychological"],
					"coverImage": {"medium": "https://example.com/m.jpg", "large": "https://example.com/m_large.jpg"},
					"synonyms": ["Monster"],
					"idMal": 1
				}
			}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	m, err := c.GetManga(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetManga failed: %v", err)
	}
	if m.Titles[0].Text != "Monster" {
		t.Fatalf("expected title Monster, got %s", m.Titles[0].Text)
	}
	if m.ChapterCount != 162 {
		t.Fatalf("expected 162 chapters, got %d", m.ChapterCount)
	}
	if m.Status != "finished" {
		t.Fatalf("expected status finished, got %s", m.Status)
	}
}

func TestSearchManga(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"Page": {
					"Media": [
						{"id": 1, "title": {"romaji": "Monster", "english": "Monster", "native": "モンスター"}, "description": "A thriller.", "startDate": {"year": 1994, "month": 12, "day": 5}, "endDate": {"year": 2001, "month": 12, "day": 20}, "status": "FINISHED", "format": "MANGA", "chapters": 162, "volumes": 18, "averageScore": 91, "meanScore": 91, "popularity": 5000, "genres": ["Mystery"], "coverImage": {"medium": "https://example.com/m.jpg", "large": "https://example.com/m_large.jpg"}, "synonyms": ["Monster"], "idMal": 1}
					]
				}
			}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	results, err := c.SearchManga(context.Background(), "monster", types.SearchOpts{Limit: 10})
	if err != nil {
		t.Fatalf("SearchManga failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Titles[0].Text != "Monster" {
		t.Fatalf("expected title Monster, got %s", results[0].Titles[0].Text)
	}
}

// TestRelationCarriesMalIDAndFormat pins that the query selects idMal/type/
// format on relation nodes and the mapping carries them. Consumers fetch a
// relation target from MAL by its idMal, and tell a light novel from a manga
// by format — TargetID is an AniList id and TargetType is "MANGA" for both.
func TestRelationCarriesMalIDAndFormat(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		r.Body.Read(b)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data": {"Media": {
			"id": 1, "title": {"romaji": "Konosuba"}, "status": "FINISHED", "format": "TV",
			"relations": {"edges": [
				{"relationType": "SEQUEL", "node": {"id": 21202, "idMal": 32380, "type": "ANIME", "format": "TV", "title": {"romaji": "S2"}}},
				{"relationType": "SOURCE", "node": {"id": 86633, "idMal": 44286, "type": "MANGA", "format": "NOVEL", "title": {"romaji": "LN"}}},
				{"relationType": "OTHER", "node": {"id": 999, "title": {"romaji": "NoIDs"}}}
			]},
			"externalLinks": []
		}}}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	m, err := c.Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	for _, field := range []string{"idMal", "format"} {
		if !strings.Contains(gotBody, field) {
			t.Fatalf("relation node selection must request %s, got query %q", field, gotBody)
		}
	}
	if len(m.Relations) != 3 {
		t.Fatalf("expected 3 relations, got %d", len(m.Relations))
	}
	if got := m.Relations[0]; got.TargetMalID != "32380" || got.TargetType != "anime" || got.TargetFormat != "TV" {
		t.Fatalf("anime relation mismatch: %+v", got)
	}
	if got := m.Relations[1]; got.TargetMalID != "44286" || got.TargetType != "manga" || got.TargetFormat != "NOVEL" {
		t.Fatalf("light-novel relation mismatch: %+v", got)
	}
	if got := m.Relations[2]; got.TargetMalID != "" || got.TargetFormat != "" {
		t.Fatalf("relation without idMal/format must stay empty: %+v", got)
	}
}
