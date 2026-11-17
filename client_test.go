package datami

import "testing"

func TestNewClient(t *testing.T) {
	cfg := Config{}
	c, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if c == nil {
		t.Fatal("client nil")
	}
	if c.MAL == nil {
		t.Fatal("MAL client nil")
	}
	if c.AniList == nil {
		t.Fatal("AniList client nil")
	}
	if c.AniDB == nil {
		t.Fatal("AniDB client nil")
	}
	if c.Kitsu == nil {
		t.Fatal("Kitsu client nil")
	}
	if c.Hikka == nil {
		t.Fatal("Hikka client nil")
	}
	if c.Kaize == nil {
		t.Fatal("Kaize client nil")
	}
	if c.Nautiljon == nil {
		t.Fatal("Nautiljon client nil")
	}
	if c.Otakotaku == nil {
		t.Fatal("Otakotaku client nil")
	}
}

func TestNewClientAniDBConfigured(t *testing.T) {
	c, err := New(Config{AniDBClientName: "x", AniDBClientVersion: 1})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if c.AniDB == nil {
		t.Fatal("AniDB client nil")
	}
}
