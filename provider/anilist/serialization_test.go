package anilist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

// TestMediaSurvivesJSONRoundTrip pins the serialization contract: a consumer
// that marshals a types.Media into a queue payload and unmarshals it on the
// other side must get back every field the provider filled. Status/MediaType
// are json:"-" on purpose (the *Field variants carry the verbatim upstream
// strings), so the canonical values must travel in those tagged fields;
// rankings/stats live only in Raw (also json:"-"), so the provider must copy
// them into the tagged Rankings/Stats columns the way the MAL provider does.
func TestMediaSurvivesJSONRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"Media": {
					"id": 1,
					"title": {"romaji": "Cowboy Bebop"},
					"status": "FINISHED",
					"format": "TV",
					"studios": {"edges": [{"isMain": true, "node": {"id": 14, "name": "Sunrise", "isAnimationStudio": true}}]},
					"externalLinks": [{"id": 9, "url": "https://example.com/s", "site": "Crunchyroll", "type": "STREAMING"}],
					"rankings": [{"rank": 1, "type": "RATED"}],
					"stats": {"scoreDistribution": [{"score": 90, "amount": 10}]}
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
	if m.Rankings == nil {
		t.Fatal("provider must fill the tagged Rankings field, not only Raw")
	}
	if m.Stats == nil {
		t.Fatal("provider must fill the tagged Stats field, not only Raw")
	}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var got types.Media
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	if len(got.Studios) != 1 || !got.Studios[0].IsAnimationStudio || !got.Studios[0].IsMain {
		t.Fatalf("studio flags lost in round trip: %+v", got.Studios)
	}
	if len(got.ExternalLinks) != 1 || got.ExternalLinks[0].URL != "https://example.com/s" {
		t.Fatalf("external links lost in round trip: %+v", got.ExternalLinks)
	}

	if got.StatusField != "FINISHED" {
		t.Fatalf("status lost in round trip: %q", got.StatusField)
	}
	if got.MediaTypeField != "TV" {
		t.Fatalf("format lost in round trip: %q", got.MediaTypeField)
	}
	if got.Rankings == nil {
		t.Fatal("rankings lost in round trip")
	}
	if got.Stats == nil {
		t.Fatal("stats lost in round trip")
	}
}
