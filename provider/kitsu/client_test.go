package kitsu

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/linkeunid/datami/types"
)

func TestGetAnime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": {
				"id": "1",
				"type": "anime",
				"attributes": {
					"slug": "cowboy-bebop",
					"synopsis": "Space bounty hunters.",
					"startDate": "1998-04-03",
					"endDate": "1998-06-26",
					"status": "finished",
					"showType": "TV",
					"episodeCount": 26,
					"episodeLength": 24,
					"averageRating": 85.5,
					"userCount": 15000,
					"favoritesCount": 500,
					"genres": ["Action", "Sci-Fi"],
					"posterImage": {"medium": "https://example.com/cb.jpg", "original": "https://example.com/cb_orig.jpg"},
					"titles": {"en": "Cowboy Bebop", "en_jp": "Cowboy Bebop", "ja_jp": "カウボーイビバップ"}
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
	if m.Score == nil || m.Score.Mean != 85.5 {
		t.Fatal("score mean mismatch")
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
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": {
				"id": "1",
				"type": "manga",
				"attributes": {
					"slug": "monster",
					"synopsis": "A thriller.",
					"startDate": "1994-12-05",
					"endDate": "2001-12-20",
					"status": "finished",
					"mangaType": "Manga",
					"chapterCount": 162,
					"volumeCount": 18,
					"averageRating": 90.5,
					"userCount": 5000,
					"genres": ["Mystery", "Psychological"],
					"posterImage": {"medium": "https://example.com/m.jpg", "original": "https://example.com/m_orig.jpg"},
					"titles": {"en": "Monster", "en_jp": "Monster", "ja_jp": "モンスター"}
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
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.Write([]byte(`{
			"data": [
				{"id": "1", "type": "manga", "attributes": {"slug": "monster", "synopsis": "A thriller.", "startDate": "1994-12-05", "endDate": "2001-12-20", "status": "finished", "mangaType": "Manga", "chapterCount": 162, "volumeCount": 18, "averageRating": 90.5, "userCount": 5000, "genres": ["Mystery"], "posterImage": {"medium": "https://example.com/m.jpg", "original": "https://example.com/m_orig.jpg"}, "titles": {"en": "Monster", "en_jp": "Monster", "ja_jp": "モンスター"}}
			}]
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
