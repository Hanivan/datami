package hikka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// Compile-time check that Client satisfies types.Provider.
var _ types.Provider = (*Client)(nil)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	return nil, helpers.Wrap(types.SourceHikka, OpGet, types.ErrNotImplemented)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = 100
	}
	page := opts.Page
	if page < 1 {
		page = 1
	}
	payload := map[string]any{
		"media_type":      []any{},
		"status":          []any{},
		"season":          []any{},
		"rating":          []any{},
		"years":           []any{},
		"genres":          []any{},
		"studios":         []any{},
		"only_translated": false,
		"sort":            []any{"score:desc"},
	}
	if query != "" {
		payload["search"] = query
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, helpers.Wrap(types.SourceHikka, OpSearch, fmt.Errorf("marshal: %w", err))
	}
	url := fmt.Sprintf("%s/anime?page=%d&size=%d", c.baseURL, page, limit)
	code, respBody, retryAfter, err := c.client.Post(ctx, OpSearch, url, body, hikkaHeaders())
	if err != nil {
		return nil, helpers.Wrap(types.SourceHikka, OpSearch, err)
	}
	return handleListResponse(code, respBody, OpSearch, retryAfter)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, helpers.Wrap(types.SourceHikka, OpGetByMalID, types.ErrNotImplemented)
}

func (c *Client) ListAll(ctx context.Context, opts types.SearchOpts) ([]types.Media, error) {
	return c.Search(ctx, "", opts)
}

func hikkaHeaders() map[string]string {
	return map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "en",
		"Content-Type":    "application/json",
		"Origin":          "https://hikka.io",
	}
}

func handleListResponse(code int, body []byte, op string, retryAfter int) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[hikkaResponse](types.SourceHikka, op, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.List))
	for _, item := range resp.List {
		if item.MalID == 0 {
			continue
		}
		results = append(results, *item.toMedia())
	}
	return results, nil
}
