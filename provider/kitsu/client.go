package kitsu

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

// Compile-time check that Client satisfies types.Provider.
var _ types.Provider = (*Client)(nil)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceKitsu, OpGet, types.ErrInvalidInput)
	}
	u := fmt.Sprintf("%s/anime/%s", c.baseURL, url.PathEscape(id))
	resp, err := helpers.FetchJSON[mediaResponse](ctx, c.client, types.SourceKitsu, OpGet, u)
	if err != nil {
		return nil, err
	}
	return resp.Data.toMedia(), nil
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, helpers.Wrap(types.SourceKitsu, OpSearch, types.ErrInvalidInput)
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime?filter[text]=%s&page[limit]=%d", c.baseURL, url.QueryEscape(query), limit)
	code, body, retryAfter, err := c.client.Get(ctx, OpSearch, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceKitsu, OpSearch, err)
	}
	return handleSearchResponse(code, body, retryAfter)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, helpers.Wrap(types.SourceKitsu, OpGetByMalID, types.ErrNotFound)
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceKitsu, OpGetManga, types.ErrInvalidInput)
	}
	u := fmt.Sprintf("%s/manga/%s", c.baseURL, url.PathEscape(id))
	resp, err := helpers.FetchJSON[mediaResponse](ctx, c.client, types.SourceKitsu, OpGetManga, u)
	if err != nil {
		return nil, err
	}
	return resp.Data.toMedia(), nil
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, helpers.Wrap(types.SourceKitsu, OpSearchManga, types.ErrInvalidInput)
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/manga?filter[text]=%s&page[limit]=%d", c.baseURL, url.QueryEscape(query), limit)
	code, body, retryAfter, err := c.client.Get(ctx, OpSearchManga, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceKitsu, OpSearchManga, err)
	}
	return handleSearchResponse(code, body, retryAfter)
}

func handleSearchResponse(code int, body []byte, retryAfter int) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[searchResponse](types.SourceKitsu, OpSearch, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, a := range resp.Data {
		results = append(results, *a.toMedia())
	}
	return results, nil
}
