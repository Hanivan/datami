package anidb

import (
	"context"
	"encoding/xml"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

var fixedNow = time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)

func loadAnime(t *testing.T, name string) anime {
	t.Helper()
	var a anime
	if err := xml.Unmarshal([]byte(readFile(t, name)), &a); err != nil {
		t.Fatal(err)
	}
	return a
}

func fullMedia(t *testing.T) *types.Media {
	a := loadAnime(t, "anime_full.xml")
	return a.toMedia(fixedNow)
}

func TestToMediaTitles(t *testing.T) {
	m := fullMedia(t)
	want := []types.Title{
		{Text: "Seikai no Monshou", Language: "x-jat", Kind: types.KindCanonical},
		{Text: "Crest of the Stars", Language: "en", Kind: types.KindEnglish},
		{Text: "星界の紋章", Language: "ja", Kind: types.KindNative},
		{Text: "Seikai no Monshou", Language: "x-jat", Kind: types.KindRomaji},
		{Text: "Seikai no Monsho", Language: "en", Kind: types.KindSynonym},
		{Text: "CotS", Language: "x-jat", Kind: types.KindShort},
	}
	if len(m.Titles) != len(want) {
		t.Fatalf("titles = %+v", m.Titles)
	}
	for i, w := range want {
		if m.Titles[i] != w {
			t.Errorf("Titles[%d] = %+v, want %+v", i, m.Titles[i], w)
		}
	}
}

func TestToMediaCore(t *testing.T) {
	m := fullMedia(t)
	if m.MediaType != types.TypeTV {
		t.Errorf("MediaType = %q", m.MediaType)
	}
	if m.EpisodeCount != 26 {
		t.Errorf("EpisodeCount = %d", m.EpisodeCount)
	}
	if m.Score == nil || m.Score.Mean != 8.5 {
		t.Errorf("Score = %+v", m.Score)
	}
	img := "https://cdn.anidb.net/images/main/440.jpg"
	if m.Images[types.SizeOriginal].URL != img || m.Images[types.SizeLarge].URL != img {
		t.Errorf("Images = %+v", m.Images)
	}
	if len(m.Studios) != 1 || m.Studios[0].Name != "Sunrise" || m.Studios[0].ID != "3313" {
		t.Errorf("Studios = %+v", m.Studios)
	}
	if m.Status != types.StatusFinished {
		t.Errorf("Status = %q", m.Status)
	}
	if m.Synopsis != "Space opera." {
		t.Errorf("Synopsis = %q", m.Synopsis)
	}
}

func TestToMediaGenresWeight(t *testing.T) {
	m := fullMedia(t)
	if len(m.Genres) != 2 || m.Genres[0] != "military" || m.Genres[1] != "space" {
		t.Fatalf("Genres = %v", m.Genres)
	}
}

func TestToMediaRelations(t *testing.T) {
	m := fullMedia(t)
	want := []types.Relation{
		{TargetID: "4", Kind: types.RelationSequel, TargetTitle: "Seikai no Senki"},
		{TargetID: "6", Kind: types.RelationAlternative, TargetTitle: "Seikai no Danshou: Tanjou"},
		{TargetID: "1623", Kind: types.RelationParentStory, TargetTitle: "Seikai no Monshou Tokubetsuhen"},
	}
	if len(m.Relations) != len(want) {
		t.Fatalf("Relations = %+v", m.Relations)
	}
	for i, w := range want {
		r := m.Relations[i]
		if r.TargetID != w.TargetID || r.Kind != w.Kind || r.TargetTitle != w.TargetTitle || r.TargetSource != types.SourceAniDB {
			t.Errorf("Relations[%d] = %+v, want %+v", i, r, w)
		}
	}
}

func TestToMediaExternalIDs(t *testing.T) {
	m := fullMedia(t)
	want := []types.ExternalID{
		{Source: types.SourceAniDB, ID: "1", ResourceType: types.ResourceAnime},
		{Source: types.SourceMAL, ID: "1", ResourceType: types.ResourceAnime},
		{Source: types.SourceAnimeNewsNetwork, ID: "13", ResourceType: types.ResourceAnime},
	}
	have := map[types.ExternalID]bool{}
	for _, e := range m.ExternalIDs {
		have[e] = true
		if e.ID == "777" {
			t.Errorf("type 99 leaked: %+v", e)
		}
	}
	for _, w := range want {
		if !have[w] {
			t.Errorf("missing %+v in %+v", w, m.ExternalIDs)
		}
	}
	if len(m.ExternalIDs) != len(want) {
		t.Errorf("ExternalIDs = %+v", m.ExternalIDs)
	}
	if m.Raw == nil {
		t.Fatal("Raw is nil")
	}
}

func TestStatusDerivation(t *testing.T) {
	cases := []struct {
		name, start, end string
		want             types.MediaStatus
	}{
		{"end past", "1998-04-03", "1998-06-26", types.StatusFinished},
		{"start future", "2027-01-05", "", types.StatusPlanned},
		{"start past no end", "2026-10-01", "", types.StatusOngoing},
		{"start past end future", "2026-10-01", "2027-03-20", types.StatusOngoing},
		{"no dates", "", "", ""},
	}
	for _, c := range cases {
		a := anime{StartDate: c.start, EndDate: c.end}
		if got := a.toMedia(fixedNow).Status; got != c.want {
			t.Errorf("%s: Status = %q, want %q", c.name, got, c.want)
		}
	}
	if got := loadAnime(t, "anime_partial.xml").toMedia(fixedNow).Status; got != types.StatusOngoing {
		t.Errorf("partial: Status = %q, want ongoing", got)
	}
}

func TestGetMapsFullFixture(t *testing.T) {
	srv := serveBody(readFile(t, "anime_full.xml"))
	defer srv.Close()
	c := testClient(srv.URL)
	c.now = func() time.Time { return fixedNow }
	m, err := c.Get(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if m.Titles[0].Kind != types.KindCanonical || m.Status != types.StatusFinished {
		t.Fatalf("Titles[0] = %+v, Status = %q", m.Titles[0], m.Status)
	}
}
