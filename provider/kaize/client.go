package kaize

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

const (
	OpGet        = "get"
	OpSearch     = "search"
	OpGetByMalID = "getbymalid"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceKaize, Op: OpGet, Err: types.ErrNotImplemented}
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	page := opts.Page
	if page < 1 {
		page = 1
	}
	url := fmt.Sprintf("%s/anime/top?page=%d", c.baseURL, page)
	code, body, retryAfter, err := c.client.Get(ctx, OpSearch, url)
	if err != nil {
		return nil, &types.Error{Source: types.SourceKaize, Op: OpSearch, Err: err}
	}
	return handleListResponse(code, body, OpSearch, retryAfter)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceKaize, Op: OpGetByMalID, Err: types.ErrNotImplemented}
}

func (c *Client) ListAll(ctx context.Context, opts types.SearchOpts) ([]types.Media, error) {
	return c.Search(ctx, "", opts)
}

func handleListResponse(code int, body []byte, op string, retryAfter int) ([]types.Media, error) {
	if code != http.StatusOK {
		return nil, helpers.StatusError(types.SourceKaize, op, code, retryAfter)
	}
	items, err := parseKaizeList(body)
	if err != nil {
		return nil, &types.Error{Source: types.SourceKaize, Op: op, Err: fmt.Errorf("parse: %w", err)}
	}
	results := make([]types.Media, 0, len(items))
	for _, item := range items {
		results = append(results, *item.toMedia())
	}
	return results, nil
}
