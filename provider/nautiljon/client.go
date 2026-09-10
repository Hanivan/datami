package nautiljon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/linkeunid/datami/types"
)

const (
	OpGet        = "get"
	OpSearch     = "search"
	OpGetByMalID = "getbymalid"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceNautiljon, Op: OpGet, Err: types.ErrNotImplemented}
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	page := opts.Page
	if page < 1 {
		page = 1
	}
	url := fmt.Sprintf("%s/animes/?dbt=%d", c.baseURL, (page-1)*15)
	code, body, err := c.client.Get(ctx, url)
	if err != nil {
		return nil, &types.Error{Source: types.SourceNautiljon, Op: OpSearch, Err: err}
	}
	return handleListResponse(code, body, "search")
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceNautiljon, Op: OpGetByMalID, Err: types.ErrNotImplemented}
}

func (c *Client) ListAll(ctx context.Context, opts types.SearchOpts) ([]types.Media, error) {
	return c.Search(ctx, "", opts)
}

func handleListResponse(code int, body []byte, op string) ([]types.Media, error) {
	switch code {
	case http.StatusOK:
		items, err := parseNautiljonList(body)
		if err != nil {
			return nil, &types.Error{Source: types.SourceNautiljon, Op: op, Err: fmt.Errorf("parse: %w", err)}
		}
		results := make([]types.Media, 0, len(items))
		for _, item := range items {
			results = append(results, *item.toMedia())
		}
		return results, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceNautiljon, Op: op, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceNautiljon, Op: op, Err: types.ErrRateLimited}
	default:
		return nil, &types.Error{Source: types.SourceNautiljon, Op: op, Err: fmt.Errorf("status %d", code)}
	}
}
