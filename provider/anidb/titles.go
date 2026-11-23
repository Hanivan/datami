package anidb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

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
