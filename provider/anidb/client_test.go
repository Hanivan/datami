package anidb

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

func TestGetAnime(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
		<anime id="1" xslt="true">
			<titles>
				<title xml:lang="en" type="official">Cowboy Bebop</title>
				<title xml:lang="ja" type="official">カウボーイビバップ</title>
			</titles>
			<description>Space bounty hunters.</description>
			<episodes>26</episodes>
			<startdate>1998-04-03</startdate>
			<enddate>1998-06-26</enddate>
			<type>TV Series</type>
			<picture>cb.jpg</picture>
		</anime>`))
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

func TestGetMangaNotImplemented(t *testing.T) {
	c := NewWithTestServer("http://dummy", Config{Timeout: 5 * time.Second})

	_, err := c.GetManga(context.Background(), "1")
	if !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}

	_, err = c.SearchManga(context.Background(), "test", types.SearchOpts{})
	if !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}

	_, err = c.GetMangaByMalID(context.Background(), "1")
	if !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}
