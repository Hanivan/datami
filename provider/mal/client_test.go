package mal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

func TestGetAnime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": 1,
			"title": "Cowboy Bebop",
			"main_picture": {"medium": "https://example.com/cb.jpg", "large": "https://example.com/cb_large.jpg"},
			"synopsis": "Space bounty hunters.",
			"start_date": "1998-04-03",
			"end_date": "1998-06-26",
			"status": "finished_airing",
			"media_type": "tv",
			"num_episodes": 26,
			"average_episode_duration": 24,
			"mean": 8.78,
			"rank": 40,
			"popularity": 30,
			"genres": [{"id": 1, "name": "Action"}, {"id": 2, "name": "Sci-Fi"}],
			"studios": [{"id": 14, "name": "Sunrise"}],
			"alternative_titles": {"en": "Cowboy Bebop", "ja": "カウボーイビバップ", "synonyms": []}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{
		ClientID: "test",
		Timeout:  5 * time.Second,
	})
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
	if m.Score == nil || m.Score.Mean != 8.78 {
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
		if got := r.Header.Get("X-MAL-CLIENT-ID"); got != "test" {
			t.Errorf("expected X-MAL-CLIENT-ID=test, got %q", got)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{
		ClientID: "test",
		Timeout:  5 * time.Second,
	})
	_, err := c.Get(context.Background(), "999999")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestGetManga(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": 1,
			"title": "Monster",
			"main_picture": {"medium": "https://example.com/m.jpg", "large": "https://example.com/m_large.jpg"},
			"synopsis": "A thriller.",
			"start_date": "1994-12-05",
			"end_date": "2001-12-20",
			"status": "finished_publishing",
			"media_type": "manga",
			"num_chapters": 162,
			"num_volumes": 18,
			"mean": 9.12,
			"rank": 10,
			"popularity": 50,
			"genres": [{"id": 1, "name": "Mystery"}, {"id": 2, "name": "Psychological"}],
			"alternative_titles": {"en": "Monster", "ja": "モンスター", "synonyms": []},
			"related_manga": [],
			"related_anime": []
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
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
	if m.Score == nil || m.Score.Mean != 9.12 {
		t.Fatal("score mean mismatch")
	}
}

func TestSearchManga(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": [
				{"node": {"id": 1, "title": "Monster", "main_picture": {"medium": "https://example.com/m.jpg", "large": "https://example.com/m_large.jpg"}, "synopsis": "A thriller.", "status": "finished_publishing", "media_type": "manga", "num_chapters": 162, "mean": 9.12, "rank": 10, "popularity": 50, "genres": [{"id": 1, "name": "Mystery"}], "alternative_titles": {"en": "Monster", "ja": "モンスター", "synonyms": []}}}
			],
			"paging": {}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
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

func TestGetSeasonal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": [
				{"node": {"id": 49233, "title": "Youjo Senki II", "main_picture": {"medium": "https://example.com/ys.jpg", "large": "https://example.com/ys_large.jpg"}, "synopsis": "Sequel.", "status": "currently_airing", "media_type": "tv", "num_episodes": 12, "mean": 8.5, "rank": 100, "popularity": 200, "genres": [{"id": 1, "name": "Action"}], "alternative_titles": {"en": "Saga of Tanya the Evil II", "ja": "幼女戦記II", "synonyms": []}}}
			],
			"paging": {}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
	results, err := c.GetSeasonal(context.Background(), 2026, "winter", types.SearchOpts{Limit: 10})
	if err != nil {
		t.Fatalf("GetSeasonal failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Titles[0].Text != "Youjo Senki II" {
		t.Fatalf("expected title Youjo Senki II, got %s", results[0].Titles[0].Text)
	}
}

func TestGetRanking(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": [
				{"node": {"id": 1, "title": "Monster", "main_picture": {"medium": "https://example.com/m.jpg", "large": "https://example.com/m_large.jpg"}, "synopsis": "Top manga.", "status": "finished_publishing", "media_type": "manga", "num_chapters": 162, "mean": 9.12, "rank": 1, "popularity": 50, "genres": [{"id": 1, "name": "Mystery"}], "alternative_titles": {"en": "Monster", "ja": "モンスター", "synonyms": []}}, "ranking": {"rank": 1}}
			],
			"paging": {}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
	results, err := c.GetMangaRanking(context.Background(), "manga", types.SearchOpts{Limit: 10})
	if err != nil {
		t.Fatalf("GetMangaRanking failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Titles[0].Text != "Monster" {
		t.Fatalf("expected title Monster, got %s", results[0].Titles[0].Text)
	}
}

func TestGetAnimeCharacters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": [
				{"node": {"id": 1, "first_name": "Kenjiro", "last_name": "Tsuda", "main_picture": {"medium": "https://example.com/tsuda.jpg"}, "biography": "A voice actor.", "num_favorites": 1000}, "role": "Main"}
			],
			"paging": {}
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
	chars, err := c.GetAnimeCharacters(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetAnimeCharacters failed: %v", err)
	}
	if len(chars) != 1 {
		t.Fatalf("expected 1 character, got %d", len(chars))
	}
	if chars[0].ID != "1" {
		t.Fatalf("expected ID 1, got %s", chars[0].ID)
	}
}

func TestGetCharacter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": 137825,
			"first_name": "Light",
			"last_name": "Yagami",
			"main_picture": {"medium": "https://example.com/light.jpg"},
			"biography": "A genius student.",
			"num_favorites": 5000
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
	ch, err := c.GetCharacter(context.Background(), "137825")
	if err != nil {
		t.Fatalf("GetCharacter failed: %v", err)
	}
	if ch.ID != "137825" {
		t.Fatalf("expected ID 137825, got %s", ch.ID)
	}
	if ch.Biography != "A genius student." {
		t.Fatalf("biography mismatch: %s", ch.Biography)
	}
}

func TestGetPerson(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": 1870,
			"first_name": "Hayao",
			"last_name": "Miyazaki",
			"main_picture": {"medium": "https://example.com/miyazaki.jpg"},
			"birthday": "1941-01-05",
			"num_favorites": 10000,
			"more": "Legendary director."
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
	p, err := c.GetPerson(context.Background(), "1870")
	if err != nil {
		t.Fatalf("GetPerson failed: %v", err)
	}
	if p.ID != "1870" {
		t.Fatalf("expected ID 1870, got %s", p.ID)
	}
	if p.Name != "Hayao Miyazaki" {
		t.Fatalf("expected name Hayao Miyazaki, got %s", p.Name)
	}
	if p.Birthday != "1941-01-05" {
		t.Fatalf("birthday mismatch: %s", p.Birthday)
	}
}

// TestRelationTargetTypeIsNodeMediaType pins the two halves of the relation
// contract: the request asks for the nested media_type selector, and the
// mapping carries that verbatim node media_type onto TargetType. A bare
// related_anime request returns nodes without media_type, which would make
// every anime target indistinguishable from a manga one downstream.
func TestRelationTargetTypeIsNodeMediaType(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": 30831,
			"title": "Konosuba",
			"media_type": "tv",
			"status": "finished_airing",
			"related_anime": [
				{"node": {"id": 32380, "title": "S2", "media_type": "tv"}, "relation_type": "sequel"},
				{"node": {"id": 51958, "title": "OVA", "media_type": "ova"}, "relation_type": "side_story"}
			],
			"related_manga": [
				{"node": {"id": 32937, "title": "Note", "media_type": "manga"}, "relation_type": "other"},
				{"node": {"id": 44286, "title": "LN", "media_type": "light_novel"}, "relation_type": "source"},
				{"node": {"id": 77, "title": "NoType"}, "relation_type": "other"}
			]
		}`))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{ClientID: "test", Timeout: 5 * time.Second})
	m, err := c.Get(context.Background(), "30831")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !strings.Contains(gotQuery, "related_anime{media_type}") || !strings.Contains(gotQuery, "related_manga{media_type}") {
		t.Fatalf("request must select nested media_type, got query %q", gotQuery)
	}
	want := map[string]string{"32380": "tv", "51958": "ova", "32937": "manga", "44286": "light_novel", "77": "manga"}
	if len(m.Relations) != len(want) {
		t.Fatalf("expected %d relations, got %d", len(want), len(m.Relations))
	}
	for _, rel := range m.Relations {
		if w, ok := want[rel.TargetID]; !ok || rel.TargetType != w {
			t.Fatalf("relation %s: TargetType = %q, want %q", rel.TargetID, rel.TargetType, w)
		}
	}
}
