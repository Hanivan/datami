package otakotaku

import (
	"context"
	"fmt"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// Compile-time check that Client satisfies types.Provider.
var _ types.Provider = (*Client)(nil)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceOtakotaku, OpGet, types.ErrInvalidInput)
	}
	url := fmt.Sprintf("%s/api/anime/view/%s/placeholder", c.baseURL, id)
	code, body, retryAfter, err := c.client.Get(ctx, OpGet, url)
	if err != nil {
		return nil, helpers.Wrap(types.SourceOtakotaku, OpGet, err)
	}
	return handleGetResponse(code, body, OpGet, retryAfter)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	return nil, helpers.Wrap(types.SourceOtakotaku, OpSearch, types.ErrNotImplemented)
}

// GetByMalID is not supported: OtakOtaku has no MAL-id lookup, and its own ids
// are a separate namespace. Use ListAll and match ExternalIDs instead.
func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return nil, helpers.Wrap(types.SourceOtakotaku, OpGetByMalID, types.ErrNotImplemented)
}

func (c *Client) ListAll(ctx context.Context, opts types.SearchOpts) ([]types.Media, error) {
	id := opts.Page
	if id < 1 {
		id = 1
	}
	url := fmt.Sprintf("%s/api/anime/view/%d/placeholder", c.baseURL, id)
	code, body, retryAfter, err := c.client.Get(ctx, OpListAll, url)
	if err != nil {
		return nil, helpers.Wrap(types.SourceOtakotaku, OpListAll, err)
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
