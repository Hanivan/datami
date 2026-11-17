package anidb

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (f *fakeClock) now() time.Time { f.mu.Lock(); defer f.mu.Unlock(); return f.t }
func (f *fakeClock) add(d time.Duration) {
	f.mu.Lock()
	f.t = f.t.Add(d)
	f.mu.Unlock()
}

func gzFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/anime-titles.xml")
	if err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	zw.Write(raw)
	zw.Close()
	return b.Bytes()
}

// dumpServer serves the gzipped fixture, or 500 while fail is set.
func dumpServer(t *testing.T, fail *atomic.Bool) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	body := gzFixture(t)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if fail != nil && fail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// titlesClient uses a fake clock and an unthrottled httpx client so
// multi-download tests do not wait on the 1 req/2s limiter.
func titlesClient(url, dir string) (*Client, *fakeClock) {
	c := NewWithTestServer("http://unused", url, Config{CacheDir: dir})
	c.client = httpx.New(httpx.Config{Timeout: 5 * time.Second, MaxRetries: 1})
	clk := &fakeClock{t: time.Now()}
	c.now = clk.now
	return c, clk
}

func writeCache(t *testing.T, dir string, mtime time.Time) {
	t.Helper()
	p := filepath.Join(dir, titlesFile)
	if err := os.WriteFile(p, gzFixture(t), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func mustSearch(t *testing.T, c *Client, q string, opts types.SearchOpts) []types.Media {
	t.Helper()
	res, err := c.Search(context.Background(), q, opts)
	if err != nil {
		t.Fatalf("Search(%q): %v", q, err)
	}
	if len(res) == 0 {
		t.Fatalf("Search(%q): no results", q)
	}
	return res
}

func checkHits(t *testing.T, hits *atomic.Int32, want int32) {
	t.Helper()
	if got := hits.Load(); got != want {
		t.Fatalf("hits = %d, want %d", got, want)
	}
}

func TestSearchDownloadsWhenMissing(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	dir := t.TempDir()
	c, _ := titlesClient(srv.URL, dir)
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 1)
	if _, err := os.Stat(filepath.Join(dir, titlesFile)); err != nil {
		t.Fatalf("cache file: %v", err)
	}
}

func TestSearchUsesFreshFile(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	dir := t.TempDir()
	c, clk := titlesClient(srv.URL, dir)
	writeCache(t, dir, clk.now())
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 0)
}

func TestSearchRefreshesOldFile(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	dir := t.TempDir()
	c, clk := titlesClient(srv.URL, dir)
	writeCache(t, dir, clk.now().Add(-25*time.Hour))
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 1)
}

func TestSearchRefreshesInProcess(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	c, clk := titlesClient(srv.URL, t.TempDir())
	mustSearch(t, c, "bebop", types.SearchOpts{})
	clk.add(25 * time.Hour)
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 2)
}

func TestSearchStaleFallback(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	srv, hits := dumpServer(t, &fail)
	dir := t.TempDir()
	c, clk := titlesClient(srv.URL, dir)
	writeCache(t, dir, clk.now().Add(-25*time.Hour))
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 1)
	clk.add(30 * time.Minute)
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 1)
	clk.add(31 * time.Minute)
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 2)
}

func TestSearchNoFileDownloadFails(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	srv, hits := dumpServer(t, &fail)
	c, clk := titlesClient(srv.URL, t.TempDir())
	for i := 0; i < 2; i++ {
		if _, err := c.Search(context.Background(), "bebop", types.SearchOpts{}); !errors.Is(err, types.ErrUnavailable) {
			t.Fatalf("Search #%d err = %v, want ErrUnavailable", i, err)
		}
		clk.add(30 * time.Minute)
	}
	checkHits(t, hits, 1)
}

func TestSearchOrderingAndLimit(t *testing.T) {
	srv, _ := dumpServer(t, nil)
	c, _ := titlesClient(srv.URL, t.TempDir())
	res := mustSearch(t, c, "cowboy bebop", types.SearchOpts{})
	if len(res) != 2 {
		t.Fatalf("len = %d, want 2 (deduped by aid)", len(res))
	}
	want := []string{"23", "5"}
	for i, m := range res {
		ext := m.ExternalIDs[0]
		if ext.Source != types.SourceAniDB || ext.ID != want[i] {
			t.Fatalf("res[%d] ext = %+v, want aid %s", i, ext, want[i])
		}
		if len(m.Titles) == 0 || m.Titles[0].Kind != types.KindCanonical {
			t.Fatalf("res[%d] titles = %+v", i, m.Titles)
		}
	}
	if res := mustSearch(t, c, "cowboy bebop", types.SearchOpts{Limit: 1}); len(res) != 1 {
		t.Fatalf("Limit 1: len = %d", len(res))
	}
	// Generated dump with more matches than the default cap.
	var gen strings.Builder
	gen.WriteString("<animetitles>")
	for i := 1; i <= httpx.DefaultSearchLimit+5; i++ {
		fmt.Fprintf(&gen, `<anime aid="%d"><title xml:lang="x-jat" type="main">Gen Title %d</title></anime>`, i, i)
	}
	gen.WriteString("</animetitles>")
	if err := c.LoadTitles(strings.NewReader(gen.String())); err != nil {
		t.Fatal(err)
	}
	all := mustSearch(t, c, "gen title", types.SearchOpts{})
	if len(all) != httpx.DefaultSearchLimit {
		t.Fatalf("default limit: len = %d, want %d", len(all), httpx.DefaultSearchLimit)
	}
	seen := map[string]bool{}
	for _, m := range all {
		id := m.ExternalIDs[0].ID
		if seen[id] {
			t.Fatalf("duplicate aid %s", id)
		}
		seen[id] = true
	}
}

func TestSearchNormalization(t *testing.T) {
	srv, _ := dumpServer(t, nil)
	c, _ := titlesClient(srv.URL, t.TempDir())
	for _, q := range []string{"re zero", "RE:ZERO"} {
		res := mustSearch(t, c, q, types.SearchOpts{})
		if res[0].ExternalIDs[0].ID != "11370" {
			t.Fatalf("Search(%q)[0] = %+v", q, res[0].ExternalIDs[0])
		}
	}
}

func TestSearchConcurrentSingleDownload(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	c, _ := titlesClient(srv.URL, t.TempDir())
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Search(context.Background(), "bebop", types.SearchOpts{}); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	checkHits(t, hits, 1)
}

func TestSearchUnwritableCacheDir(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var metricErrs atomic.Int32
	c, clk := titlesClient(srv.URL, filepath.Join(f, "sub"))
	c.metrics = func(source, op string, dur time.Duration, err error) {
		if err != nil {
			metricErrs.Add(1)
		}
	}
	mustSearch(t, c, "bebop", types.SearchOpts{})
	clk.add(time.Hour)
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 1)
	if metricErrs.Load() == 0 {
		t.Fatal("write error not reported via Metrics")
	}
}

func TestLoadTitlesNoNetwork(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	srv, hits := dumpServer(t, &fail)
	c, _ := titlesClient(srv.URL, t.TempDir())
	if err := c.LoadTitles(bytes.NewReader(gzFixture(t))); err != nil {
		t.Fatal(err)
	}
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 0)
	// Plain XML also loads.
	raw, _ := os.ReadFile("testdata/anime-titles.xml")
	if err := c.LoadTitles(bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	c, _ := titlesClient(srv.URL, t.TempDir())
	if _, err := c.Search(context.Background(), "", types.SearchOpts{}); !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("err = %v", err)
	}
	if _, err := Unconfigured().Search(context.Background(), "bebop", types.SearchOpts{}); !errors.Is(err, types.ErrInvalidInput) {
		t.Fatalf("unconfigured err = %v", err)
	}
	checkHits(t, hits, 0)
}

func TestSearchFutureMtimeNotFresh(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	dir := t.TempDir()
	c, clk := titlesClient(srv.URL, dir)
	writeCache(t, dir, clk.now().Add(time.Hour))
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 1)
}

func TestSearchCancelledCtxNoLockout(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	c, _ := titlesClient(srv.URL, t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Search(ctx, "bebop", types.SearchOpts{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 1)
}

func TestSearchCancelMidTransferSetsCooldown(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	c, clk := titlesClient(srv.URL, t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := c.Search(ctx, "bebop", types.SearchOpts{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	clk.add(30 * time.Minute)
	if _, err := c.Search(context.Background(), "bebop", types.SearchOpts{}); err == nil {
		t.Fatal("second Search: want cooldown error")
	}
	checkHits(t, &hits, 1)
}

// Uses the real 1 req/2s client: a ctx deadline shorter than the 2s token
// wait makes limiter.Wait refuse with no request sent.
func TestSearchLimiterRefusalNoCooldown(t *testing.T) {
	srv, hits := dumpServer(t, nil)
	c := NewWithTestServer("http://unused", srv.URL, Config{CacheDir: t.TempDir()})
	clk := &fakeClock{t: time.Now()}
	c.now = clk.now
	// Consume the only token.
	if _, _, err := c.client.Get(context.Background(), srv.URL); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := c.Search(ctx, "bebop", types.SearchOpts{})
	if !errors.Is(err, httpx.ErrLimiterWait) || ctx.Err() != nil {
		t.Fatalf("err = %v, ctx.Err = %v, want early ErrLimiterWait", err, ctx.Err())
	}
	checkHits(t, hits, 1)
	c.client = httpx.New(httpx.Config{Timeout: 5 * time.Second, MaxRetries: 1})
	mustSearch(t, c, "bebop", types.SearchOpts{})
	checkHits(t, hits, 2)
}

func TestSearchWaiterHonorsCtx(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(started) })
		<-release
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	var relOnce sync.Once
	unblock := func() { relOnce.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	c, _ := titlesClient(srv.URL, t.TempDir())

	first := make(chan struct{})
	go func() {
		defer close(first)
		c.Search(context.Background(), "bebop", types.SearchOpts{})
	}()
	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := c.Search(ctx, "bebop", types.SearchOpts{})
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err = %v, want DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("waiting Search ignored ctx")
	}
	unblock()
	<-first
}

func TestParseTitlesEmpty(t *testing.T) {
	if _, err := parseTitles(strings.NewReader("<animetitles></animetitles>")); err == nil {
		t.Fatal("want error for empty dump")
	}
}

func TestParseTitlesLimit(t *testing.T) {
	dump := `<animetitles><anime aid="1"><title xml:lang="x-jat" type="main">A</title></anime></animetitles>`
	if _, err := parseTitlesN(strings.NewReader(dump), int64(len(dump)-1)); !errors.Is(err, errTooLarge) {
		t.Fatalf("err = %v, want errTooLarge", err)
	}
	if _, err := parseTitlesN(strings.NewReader(dump), int64(len(dump))); err != nil {
		t.Fatalf("at limit: %v", err)
	}
}

func TestDefaultTitlesURLHTTPS(t *testing.T) {
	if !strings.HasPrefix(defaultTitlesURL, "https://") {
		t.Fatal(defaultTitlesURL)
	}
}
