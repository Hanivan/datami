package nautiljon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// Compile-time check that Client satisfies types.Provider.
var _ types.Provider = (*Client)(nil)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	return nil, helpers.Wrap(types.SourceNautiljon, OpGet, types.ErrNotImplemented)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	page := opts.Page
	if page < 1 {
		page = 1
	}
	url := fmt.Sprintf("%s/animes/?dbt=%d", c.baseURL, (page-1)*15)
	code, body, retryAfter, err := c.client.Get(ctx, OpSearch, url)
	if err != nil {
		return nil, helpers.Wrap(types.SourceNautiljon, OpSearch, err)
	}
	return handleListResponse(code, body, OpSearch, retryAfter)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, helpers.Wrap(types.SourceNautiljon, OpGetByMalID, types.ErrNotImplemented)
}

func (c *Client) ListAll(ctx context.Context, opts types.SearchOpts) ([]types.Media, error) {
	return c.Search(ctx, "", opts)
}

func handleListResponse(code int, body []byte, op string, retryAfter int) ([]types.Media, error) {
	if code != http.StatusOK {
		return nil, helpers.StatusError(types.SourceNautiljon, op, code, retryAfter)
	}
	items, err := parseNautiljonList(body)
	if err != nil {
		return nil, helpers.Wrap(types.SourceNautiljon, op, fmt.Errorf("parse: %w", err))
	}
	results := make([]types.Media, 0, len(items))
	for _, item := range items {
		results = append(results, *item.toMedia())
	}
	return results, nil
}
