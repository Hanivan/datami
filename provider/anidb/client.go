package anidb

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Hanivan/datami/types"
)

const (
	OpGet             = "get"
	OpSearch          = "search"
	OpGetByMalID      = "getbymalid"
	OpGetManga        = "getmanga"
	OpSearchManga     = "searchmanga"
	OpGetMangaByMalID = "getmangabymalid"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpGet, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s?request=anime&aid=%s", c.baseURL, url.QueryEscape(id))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpGet, Err: err}
	}
	return handleResponse(code, body, OpGet)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearch, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s?request=search&query=%s", c.baseURL, url.QueryEscape(query))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearch, Err: err}
	}
	return handleSearchResponse(code, body)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceAniDB, Op: OpGetByMalID, Err: types.ErrNotFound}
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceAniDB, Op: OpGetManga, Err: types.ErrNotImplemented}
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearchManga, Err: types.ErrNotImplemented}
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceAniDB, Op: OpGetMangaByMalID, Err: types.ErrNotImplemented}
}

func handleResponse(code int, body []byte, op string) (*types.Media, error) {
	switch code {
	case http.StatusOK:
		var a anime
		if err := xml.Unmarshal(body, &a); err != nil {
			return nil, &types.Error{Source: types.SourceAniDB, Op: op, Err: fmt.Errorf("parse: %w", err)}
		}
		return a.toMedia(), nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceAniDB, Op: op, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceAniDB, Op: op, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceAniDB, Op: op, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceAniDB, Op: op, Err: fmt.Errorf("status %d", code)}
	}
}

func handleSearchResponse(code int, body []byte) ([]types.Media, error) {
	switch code {
	case http.StatusOK:
		var resp searchResponse
		if err := xml.Unmarshal(body, &resp); err != nil {
			return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearch, Err: fmt.Errorf("parse: %w", err)}
		}
		results := make([]types.Media, 0, len(resp.Anime))
		for _, a := range resp.Anime {
			results = append(results, *a.toMedia())
		}
		return results, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearch, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearch, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearch, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceAniDB, Op: OpSearch, Err: fmt.Errorf("status %d", code)}
	}
}
