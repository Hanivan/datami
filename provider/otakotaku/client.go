package otakotaku

import (
	"context"
	"fmt"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

const (
	OpGet        = "get"
	OpSearch     = "search"
	OpGetByMalID = "getbymalid"
	OpListAll    = "listall"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpGet, Err: types.ErrInvalidInput}
	}
	url := fmt.Sprintf("%s/api/anime/view/%s/placeholder", c.baseURL, id)
	code, body, retryAfter, err := c.client.Get(ctx, OpGet, url)
	if err != nil {
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpGet, Err: err}
	}
	return handleGetResponse(code, body, OpGet, retryAfter)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpSearch, Err: types.ErrNotImplemented}
}

// GetByMalID is not supported: OtakOtaku has no MAL-id lookup, and its own ids
// are a separate namespace. Use ListAll and match ExternalIDs instead.
func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpGetByMalID, Err: types.ErrNotImplemented}
}

func (c *Client) ListAll(ctx context.Context, opts types.SearchOpts) ([]types.Media, error) {
	id := opts.Page
	if id < 1 {
		id = 1
	}
	url := fmt.Sprintf("%s/api/anime/view/%d/placeholder", c.baseURL, id)
	code, body, retryAfter, err := c.client.Get(ctx, OpListAll, url)
	if err != nil {
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpListAll, Err: err}
	}
	resp, err := handleGetResponse(code, body, OpListAll, retryAfter)
	if err != nil {
		return nil, err
	}
	return []types.Media{*resp}, nil
}

func handleGetResponse(code int, body []byte, op string, retryAfter int) (*types.Media, error) {
	resp, err := helpers.DecodeJSON[otakotakuResponse](types.SourceOtakotaku, op, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	return resp.toMedia(), nil
}
