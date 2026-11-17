package anidb

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

const sampleAnime = `<?xml version="1.0" encoding="UTF-8"?>
<anime id="1" restricted="false">
	<type>TV Series</type>
	<episodecount>26</episodecount>
	<startdate>1998-04-03</startdate>
	<enddate>1998-06-26</enddate>
	<titles>
		<title xml:lang="x-jat" type="main">Cowboy Bebop</title>
		<title xml:lang="ja" type="official">カウボーイビバップ</title>
	</titles>
	<relatedanime>
		<anime id="5" type="Side Story">Cowboy Bebop: Tengoku no Tobira</anime>
	</relatedanime>
	<description>Space bounty hunters.</description>
	<picture>cb.jpg</picture>
</anime>`

func serveBody(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
}

func testClient(url string) *Client {
	return NewWithTestServer(url, "", Config{Timeout: 5 * time.Second})
}

func TestGetAnime(t *testing.T) {
	srv := serveBody(sampleAnime)
	defer srv.Close()
	m, err := testClient(srv.URL).Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if m.Titles[0].Text != "Cowboy Bebop" || m.Titles[0].Language != "x-jat" {
		t.Fatalf("title = %+v", m.Titles[0])
	}
	if len(m.Relations) != 1 || m.Relations[0].TargetID != "5" {
		t.Fatalf("relations = %+v", m.Relations)
	}
	if m.StartDate == nil || *m.StartDate.Year != 1998 {
		t.Fatal("start date year mismatch")
	}
}

func TestNewRequiresClient(t *testing.T) {
	for _, cfg := range []Config{{}, {ClientName: "x"}} {
		if _, err := New(cfg); !errors.Is(err, types.ErrInvalidInput) {
			t.Fatalf("New(%+v) err = %v, want ErrInvalidInput", cfg, err)
		}
	}
	if _, err := New(Config{ClientName: "x", ClientVersion: 1}); err != nil {
		t.Fatalf("New valid: %v", err)
	}
}

func TestGetSendsClientParams(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		w.Write([]byte(sampleAnime))
	}))
	defer srv.Close()
	c := NewWithTestServer(srv.URL, "", Config{ClientName: "myclient", ClientVersion: 3})
	if _, err := c.Get(context.Background(), "1"); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"request": "anime", "client": "myclient", "clientver": "3", "protover": "1", "aid": "1"}
	for k, v := range want {
		if q.Get(k) != v {
			t.Fatalf("%s = %q, want %q (query %v)", k, q.Get(k), v, q)
		}
	}
}

func TestGetInvalidAidNoRequest(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
	}))
	defer srv.Close()
	c := testClient(srv.URL)
	for _, aid := range []string{"", "abc", " 1", "-1", "+1", "01"} {
		if _, err := c.Get(context.Background(), aid); !errors.Is(err, types.ErrInvalidInput) {
			t.Fatalf("Get(%q) err = %v, want ErrInvalidInput", aid, err)
		}
	}
	if hits != 0 {
		t.Fatalf("hits = %d, want 0", hits)
	}
}

func readFile(t *testing.T, name string) string {
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestErrorBodyMapping(t *testing.T) {
	cases := []struct {
		body string
		want error
	}{
		{readFile(t, "error_banned.xml"), types.ErrRateLimited},
		{"<error>\n  Banned \n</error>", types.ErrRateLimited},
		{"<error>Banned: flooding</error>", types.ErrRateLimited},
		{"<error>banned</error>", types.ErrRateLimited},
		{readFile(t, "error_notfound.xml"), types.ErrNotFound},
		{readFile(t, "error_client.xml"), types.ErrAuthFailed},
		{"<error>something new</error>", types.ErrUnavailable},
	}
	for _, tc := range cases {
		srv := serveBody(tc.body)
		_, err := testClient(srv.URL).Get(context.Background(), "1")
		srv.Close()
		if !errors.Is(err, tc.want) {
			t.Fatalf("body %q: err = %v, want %v", tc.body, err, tc.want)
		}
		var te *types.Error
		if !errors.As(err, &te) || te.Source != types.SourceAniDB || te.Op != OpGet {
			t.Fatalf("body %q: not a types.Error with anidb/get: %#v", tc.body, err)
		}
		if tc.want == types.ErrUnavailable && !strings.Contains(err.Error(), "something new") {
			t.Fatalf("err %q missing text", err)
		}
	}
}

func TestGetGzipBody(t *testing.T) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(sampleAnime))
	zw.Close()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(buf.Bytes())
	}))
	defer srv.Close()
	m, err := testClient(srv.URL).Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("Get gzip: %v", err)
	}
	if m.Titles[0].Text != "Cowboy Bebop" {
		t.Fatalf("title = %q", m.Titles[0].Text)
	}
}

func TestHTTPStatusMapping(t *testing.T) {
	cases := []struct {
		code int
		want error
	}{
		{401, types.ErrAuthFailed},
		{403, types.ErrAuthFailed},
		{404, types.ErrNotFound},
		{429, types.ErrRateLimited},
		{503, types.ErrRateLimited},
		{500, types.ErrUnavailable},
	}
	for _, tc := range cases {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.code)
		}))
		_, err := testClient(srv.URL).Get(context.Background(), "1")
		srv.Close()
		if !errors.Is(err, tc.want) {
			t.Fatalf("%d: err = %v, want %v", tc.code, err, tc.want)
		}
		if types.StatusOf(err) != tc.code {
			t.Fatalf("%d: Status = %d", tc.code, types.StatusOf(err))
		}
	}
}

func TestUnconfiguredClient(t *testing.T) {
	if _, err := Unconfigured().Get(context.Background(), "1"); !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestGetByMalIDNotImplemented(t *testing.T) {
	if _, err := testClient("http://dummy").GetByMalID(context.Background(), "1"); !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetMangaNotImplemented(t *testing.T) {
	c := testClient("http://dummy")
	if _, err := c.GetManga(context.Background(), "1"); !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("GetManga err = %v", err)
	}
	if _, err := c.SearchManga(context.Background(), "test", types.SearchOpts{}); !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("SearchManga err = %v", err)
	}
	if _, err := c.GetMangaByMalID(context.Background(), "1"); !errors.Is(err, types.ErrNotImplemented) {
		t.Fatalf("GetMangaByMalID err = %v", err)
	}
}
