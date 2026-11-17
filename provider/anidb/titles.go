package anidb

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

const (
	titlesMaxAge        = 24 * time.Hour
	titlesRetryCooldown = time.Hour
	titlesFile          = "anime-titles.xml.gz"
)

// titlesDump is the anime-titles.xml layout (format per go-anidb, see NOTICE).
type titlesDump struct {
	Anime []struct {
		AID    int          `xml:"aid,attr"`
		Titles []titleEntry `xml:"title"`
	} `xml:"anime"`
}

type titleRecord struct {
	aid   int
	raw   titles
	norms []string
}

// titlesState is the in-memory index plus refresh bookkeeping. sem is a
// 1-slot lock guarding the rest; it is held across the download so concurrent
// Searches share one request, and waiters can give up via ctx.
type titlesState struct {
	sem         chan struct{}
	index       []titleRecord // sorted by aid
	loadedAt    time.Time
	lastAttempt time.Time
	lastErr     error
}

// LoadTitles parses a gzip or plain anime-titles XML dump into the index and
// marks it loaded now, so Search uses it without network for titlesMaxAge.
func (c *Client) LoadTitles(r io.Reader) error {
	idx, err := parseTitles(r)
	if err != nil {
		return wrap(OpSearch, err)
	}
	s := &c.titles
	s.sem <- struct{}{}
	defer func() { <-s.sem }()
	s.index, s.loadedAt, s.lastErr = idx, c.now(), nil
	return nil
}

// titleIndex returns a usable index, refreshing from file or network per the
// 24h max age / 1h retry cooldown rules.
func (c *Client) titleIndex(ctx context.Context) ([]titleRecord, error) {
	s := &c.titles
	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, wrap(OpSearch, ctx.Err())
	}
	defer func() { <-s.sem }()
	now := c.now()
	if s.index != nil && now.Sub(s.loadedAt) < titlesMaxAge {
		return s.index, nil
	}
	if s.lastErr != nil && now.Sub(s.lastAttempt) < titlesRetryCooldown {
		if s.index != nil {
			return s.index, nil
		}
		return nil, s.lastErr
	}

	path := filepath.Join(c.cacheDir, titlesFile)
	fi, statErr := os.Stat(path)
	if statErr == nil && !fi.ModTime().After(now) && now.Sub(fi.ModTime()) < titlesMaxAge {
		if idx, err := readTitlesFile(path); err == nil {
			s.index, s.loadedAt, s.lastErr = idx, now, nil
			return idx, nil
		}
		// Corrupt fresh file: fall through to download.
	}

	if err := ctx.Err(); err != nil {
		// Ctx already done: nothing sent, so no cooldown.
		return nil, wrap(OpSearch, err)
	}
	data, err := c.downloadTitles(ctx)
	if errors.Is(err, httpx.ErrLimiterWait) {
		// No request sent, so no 1h cooldown.
		return nil, wrap(OpSearch, err)
	}
	// Request was sent: any failure, ctx cancel mid-transfer included, sets the
	// cooldown so a short caller timeout cannot loop re-downloads.
	s.lastAttempt = now
	var idx []titleRecord
	if err == nil {
		idx, err = parseTitles(bytes.NewReader(data))
	}
	if err == nil {
		if werr := writeTitlesFile(path, data, now); werr != nil && c.metrics != nil {
			c.metrics(types.SourceAniDB, OpSearch, 0, werr)
		}
		s.index, s.loadedAt, s.lastErr = idx, now, nil
		return idx, nil
	}

	if !errors.Is(err, types.ErrUnavailable) {
		err = fmt.Errorf("%w: %w", types.ErrUnavailable, err)
	}
	s.lastErr = wrap(OpSearch, fmt.Errorf("titles dump: %w", err))
	if statErr == nil {
		if idx, ferr := readTitlesFile(path); ferr == nil {
			// Serve the stale file; retry the download in titlesRetryCooldown.
			s.index, s.loadedAt = idx, now.Add(-titlesMaxAge+titlesRetryCooldown)
			return idx, nil
		}
	}
	if s.index != nil {
		return s.index, nil
	}
	return nil, s.lastErr
}

func (c *Client) downloadTitles(ctx context.Context) ([]byte, error) {
	code, body, err := c.client.Get(ctx, c.titlesURL)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("status %d", code)
	}
	return body, nil
}

func readTitlesFile(path string) ([]titleRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseTitles(f)
}

// writeTitlesFile writes atomically (temp + rename) and stamps mtime = now.
func writeTitlesFile(path string, data []byte, now time.Time) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, titlesFile+".tmp*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chtimes(tmp, now, now)
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
	}
	return err
}

// parseTitles streams a gzip (sniffed by magic bytes) or plain XML dump.
func parseTitles(r io.Reader) ([]titleRecord, error) {
	return parseTitlesN(r, maxTitlesDump)
}

type countReader struct {
	r io.Reader
	n int64
}

func (c *countReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// parseTitlesN is parseTitles with a cap on decompressed bytes.
func parseTitlesN(r io.Reader, max int64) ([]titleRecord, error) {
	br := bufio.NewReader(r)
	var src io.Reader = br
	if magic, _ := br.Peek(2); len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		zr, err := gzip.NewReader(br)
		if err != nil {
			return nil, fmt.Errorf("gzip: %w", err)
		}
		defer zr.Close()
		src = zr
	}
	cr := &countReader{r: io.LimitReader(src, max+1)}
	var d titlesDump
	err := xml.NewDecoder(cr).Decode(&d)
	if err == nil {
		io.Copy(io.Discard, cr) // count trailing bytes too
	}
	if cr.n > max {
		return nil, fmt.Errorf("parse titles: %w", errTooLarge)
	}
	if err != nil {
		return nil, fmt.Errorf("parse titles: %w", err)
	}
	if len(d.Anime) == 0 {
		return nil, fmt.Errorf("%w: empty titles dump", types.ErrUnavailable)
	}
	idx := make([]titleRecord, 0, len(d.Anime))
	for _, a := range d.Anime {
		r := titleRecord{aid: a.AID, raw: titles{Title: a.Titles}}
		for _, t := range a.Titles {
			r.norms = append(r.norms, normalizeTitle(t.Text))
		}
		idx = append(idx, r)
	}
	sort.Slice(idx, func(i, j int) bool { return idx[i].aid < idx[j].aid })
	return idx, nil
}

// normalizeTitle lowercases, keeps letters/digits, collapses the rest to
// single spaces.
func normalizeTitle(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(r)
		} else {
			space = true
		}
	}
	return b.String()
}

// searchIndex returns exact matches then substring matches, aid order,
// deduped, capped at limit. q must be normalized.
func searchIndex(idx []titleRecord, q string, limit int) []types.Media {
	out := []types.Media{}
	seen := map[int]bool{}
	for _, exact := range []bool{true, false} {
		for _, r := range idx {
			if len(out) >= limit {
				return out
			}
			if seen[r.aid] {
				continue
			}
			for _, n := range r.norms {
				if (exact && n == q) || (!exact && strings.Contains(n, q)) {
					seen[r.aid] = true
					out = append(out, types.Media{
						Titles: buildTitles(r.raw),
						ExternalIDs: []types.ExternalID{
							{Source: types.SourceAniDB, ID: strconv.Itoa(r.aid), ResourceType: types.ResourceAnime},
						},
					})
					break
				}
			}
		}
	}
	return out
}

func newTitlesState() titlesState {
	return titlesState{sem: make(chan struct{}, 1)}
}
