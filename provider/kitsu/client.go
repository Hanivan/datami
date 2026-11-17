package kitsu

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

const (
	OpGet         = "get"
	OpSearch      = "search"
	OpGetByMalID  = "getbymalid"
	OpGetManga    = "getmanga"
	OpSearchManga = "searchmanga"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpGet, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/anime/%s", c.baseURL, url.PathEscape(id))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpGet, Err: err}
	}
	return handleResponse(code, body, OpGet)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpSearch, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime?filter[text]=%s&page[limit]=%d", c.baseURL, url.QueryEscape(query), limit)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpSearch, Err: err}
	}
	return handleSearchResponse(code, body)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceKitsu, Op: OpGetByMalID, Err: types.ErrNotFound}
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpGetManga, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/manga/%s", c.baseURL, url.PathEscape(id))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpGetManga, Err: err}
	}
	return handleResponse(code, body, OpGetManga)
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpSearchManga, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/manga?filter[text]=%s&page[limit]=%d", c.baseURL, url.QueryEscape(query), limit)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceKitsu, Op: OpSearchManga, Err: err}
	}
	return handleSearchResponse(code, body)
}

func handleResponse(code int, body []byte, op string) (*types.Media, error) {
	resp, err := helpers.DecodeJSON[struct {
		Data anime `json:"data"`
	}](types.SourceKitsu, op, code, body)
	if err != nil {
		return nil, err
	}
	return resp.Data.toMedia(), nil
}

func handleSearchResponse(code int, body []byte) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[searchResponse](types.SourceKitsu, OpSearch, code, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, a := range resp.Data {
		results = append(results, *a.toMedia())
	}
	return results, nil
}
