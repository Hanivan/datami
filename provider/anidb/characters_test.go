package anidb

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/Hanivan/datami/types"
)

func fullChars(t *testing.T) []types.Character {
	t.Helper()
	srv := serveBody(readFile(t, "anime_full.xml"))
	defer srv.Close()
	cs, err := testClient(srv.URL).GetCharacters(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	return cs
}

func TestGetCharactersMapping(t *testing.T) {
	cs := fullChars(t)
	if len(cs) != 2 {
		t.Fatalf("len = %d, want 2", len(cs))
	}
	c := cs[0]
	want := types.Character{
		ID:        "28",
		Names:     []string{"Abriel Nei Debrusc Borl Paryun Lafiel"},
		Gender:    "female",
		Biography: "Princess of the Abh.",
		Image:     "https://cdn.anidb.net/images/main/14304.jpg",
		Roles:     []types.CharacterRole{{MediaID: "1", Role: types.RoleMain}},
		VoiceActors: []types.VoiceActor{{ID: "12", Name: "Kawasumi Ayako", Language: "ja",
			Image: "https://cdn.anidb.net/images/main/184301.jpg"}},
	}
	if !reflect.DeepEqual(c, want) {
		t.Fatalf("got %+v\nwant %+v", c, want)
	}
	if c.Tags != nil {
		t.Fatal("Tags must be nil")
	}
}

func TestGetCharactersNoSeiyuu(t *testing.T) {
	c := fullChars(t)[1]
	if c.ID != "29" || c.Roles[0].Role != types.RoleSupporting || len(c.VoiceActors) != 0 {
		t.Fatalf("got %+v", c)
	}
}

func TestGetCharactersErrors(t *testing.T) {
	srv := serveBody("<error>Banned</error>")
	defer srv.Close()
	_, err := testClient(srv.URL).GetCharacters(context.Background(), "1")
	var te *types.Error
	if !errors.Is(err, types.ErrRateLimited) || !errors.As(err, &te) || te.Op != "getcharacters" {
		t.Fatalf("err = %v", err)
	}

	var hits int32
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { atomic.AddInt32(&hits, 1) }))
	defer s2.Close()
	_, err = testClient(s2.URL).GetCharacters(context.Background(), "")
	if !errors.Is(err, types.ErrInvalidInput) || hits != 0 {
		t.Fatalf("err = %v hits = %d", err, hits)
	}
}
