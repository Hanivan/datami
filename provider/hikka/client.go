package hikka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

const (
	OpGet        = "get"
	OpSearch     = "search"
	OpGetByMalID = "getbymalid"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceHikka, Op: OpGet, Err: types.ErrNotImplemented}
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
		return nil, &types.Error{Source: types.SourceHikka, Op: OpSearch, Err: fmt.Errorf("marshal: %w", err)}
	}
	url := fmt.Sprintf("%s/anime?page=%d&size=%d", c.baseURL, page, limit)
	code, respBody, err := c.client.Post(ctx, url, body, hikkaHeaders())
	if err != nil {
		return nil, &types.Error{Source: types.SourceHikka, Op: OpSearch, Err: err}
	}
	return handleListResponse(code, respBody, OpSearch)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceHikka, Op: OpGetByMalID, Err: types.ErrNotImplemented}
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

func handleListResponse(code int, body []byte, op string) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[hikkaResponse](types.SourceHikka, op, code, body)
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
