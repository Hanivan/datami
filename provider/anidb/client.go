package anidb

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

const (
	OpNew             = "new"
	OpGet             = "get"
	OpSearch          = "search"
	OpGetByMalID      = "getbymalid"
	OpGetManga        = "getmanga"
	OpSearchManga     = "searchmanga"
	OpGetMangaByMalID = "getmangabymalid"
	OpGetCharacters   = "getcharacters"
)

func wrap(op string, err error) error {
	return &types.Error{Source: types.SourceAniDB, Op: op, Err: err}
}

// guard returns ErrInvalidInput for an Unconfigured client.
func (c *Client) guard(op string) error {
	if c.unconfigured {
		return wrap(op, fmt.Errorf("%w: anidb client name/version not configured", types.ErrInvalidInput))
	}
	return nil
}

func (c *Client) Get(ctx context.Context, aid string) (*types.Media, error) {
	a, err := c.fetchAnime(ctx, OpGet, aid)
	if err != nil {
		return nil, err
	}
	return a.toMedia(c.now()), nil
}

// Search matches query against the cached anime-titles dump (see titles.go).
// Results carry only titles and the AniDB ExternalID; call Get for detail.
func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if err := c.guard(OpSearch); err != nil {
		return nil, err
	}
	q := normalizeTitle(query)
	if q == "" {
		return nil, wrap(OpSearch, types.ErrInvalidInput)
	}
	idx, err := c.titleIndex(ctx)
	if err != nil {
		return nil, err
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = httpx.DefaultSearchLimit
	}
	return searchIndex(idx, q, limit), nil
}

func (c *Client) GetCharacters(ctx context.Context, aid string) ([]types.Character, error) {
	a, err := c.fetchAnime(ctx, OpGetCharacters, aid)
	if err != nil {
		return nil, err
	}
	return a.toCharacters(aid), nil
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if err := c.guard(OpGetByMalID); err != nil {
		return nil, err
	}
	return nil, wrap(OpGetByMalID, types.ErrNotImplemented)
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if err := c.guard(OpGetManga); err != nil {
		return nil, err
	}
	return nil, wrap(OpGetManga, types.ErrNotImplemented)
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if err := c.guard(OpSearchManga); err != nil {
		return nil, err
	}
	return nil, wrap(OpSearchManga, types.ErrNotImplemented)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if err := c.guard(OpGetMangaByMalID); err != nil {
		return nil, err
	}
	return nil, wrap(OpGetMangaByMalID, types.ErrNotImplemented)
}

// fetchAnime requests and decodes one anime by aid. Errors are *types.Error
// with the given op.
func (c *Client) fetchAnime(ctx context.Context, op, aid string) (*anime, error) {
	if err := c.guard(op); err != nil {
		return nil, err
	}
	if n, err := strconv.Atoi(aid); err != nil || n <= 0 || strconv.Itoa(n) != aid {
		return nil, wrap(op, types.ErrInvalidInput)
	}
	q := url.Values{
		"request":   {"anime"},
		"client":    {c.clientName},
		"clientver": {strconv.Itoa(c.clientVersion)},
		"protover":  {"1"},
		"aid":       {aid},
	}
	code, body, err := c.client.Get(ctx, c.apiURL+"?"+q.Encode())
	if code != http.StatusOK {
		if e := statusError(op, code); e != nil {
			return nil, e
		}
	}
	if err != nil {
		return nil, wrap(op, err)
	}
	body, err = gunzip(body)
	if err != nil {
		return nil, wrap(op, fmt.Errorf("gzip: %w", err))
	}
	if err := bodyError(body); err != nil {
		return nil, wrap(op, err)
	}
	var a anime
	if err := xml.Unmarshal(body, &a); err != nil {
		return nil, wrap(op, fmt.Errorf("parse: %w", err))
	}
	return &a, nil
}

// statusError maps a non-200 HTTP status; nil for code 0 (no response).
func statusError(op string, code int) error {
	var sentinel error
	switch {
	case code == 0:
		return nil
	case code == http.StatusNotFound:
		sentinel = types.ErrNotFound
	case code == http.StatusUnauthorized, code == http.StatusForbidden:
		sentinel = types.ErrAuthFailed
	case code == http.StatusTooManyRequests, code == http.StatusServiceUnavailable:
		sentinel = types.ErrRateLimited
	case code >= httpx.ServerErrThreshold:
		sentinel = types.ErrUnavailable
	default:
		sentinel = fmt.Errorf("%w: status %d", types.ErrUnavailable, code)
	}
	return &types.Error{Source: types.SourceAniDB, Op: op, Err: sentinel, Status: code}
}

// gunzip decompresses by magic bytes; Go only auto-decompresses when it set
// Accept-Encoding itself.
func gunzip(body []byte) ([]byte, error) {
	if len(body) < 2 || body[0] != 0x1f || body[1] != 0x8b {
		return body, nil
	}
	zr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

// bodyError maps an AniDB <error> root element (sent with HTTP 200).
func bodyError(body []byte) error {
	d := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := d.Token()
		if err != nil {
			return nil // not our concern here; Unmarshal reports parse errors
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "error" {
			return nil
		}
		var e struct {
			Text string `xml:",chardata"`
		}
		if err := d.DecodeElement(&e, &se); err != nil {
			return fmt.Errorf("%w: parse error body: %v", types.ErrUnavailable, err)
		}
		text := strings.TrimSpace(e.Text)
		lower := strings.ToLower(text)
		switch {
		case strings.HasPrefix(lower, "banned"):
			return types.ErrRateLimited
		case strings.Contains(lower, "not found"):
			return types.ErrNotFound
		case strings.Contains(lower, "client"):
			return types.ErrAuthFailed
		default:
			return fmt.Errorf("%w: %s", types.ErrUnavailable, text)
		}
	}
}
