package otakotaku

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/linkeunid/datami/types"
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
	code, body, err := c.client.Get(ctx, url)
	if err != nil {
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpGet, Err: err}
	}
	return handleGetResponse(code, body, OpGet)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpSearch, Err: types.ErrNotImplemented}
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if malID == "" {
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpGetByMalID, Err: types.ErrInvalidInput}
	}
	id, err := strconv.Atoi(malID)
	if err != nil {
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpGetByMalID, Err: types.ErrInvalidInput}
	}
	results, err := c.ListAll(ctx, types.SearchOpts{Page: id, Limit: 1})
	if err != nil {
		return nil, err
	}
	for _, r := range results {
		for _, ext := range r.ExternalIDs {
			if ext.Source == types.SourceMAL && ext.ID == malID {
				return &r, nil
			}
		}
	}
	return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpGetByMalID, Err: types.ErrNotFound}
}

func (c *Client) ListAll(ctx context.Context, opts types.SearchOpts) ([]types.Media, error) {
	id := opts.Page
	if id < 1 {
		id = 1
	}
	url := fmt.Sprintf("%s/api/anime/view/%d/placeholder", c.baseURL, id)
	code, body, err := c.client.Get(ctx, url)
	if err != nil {
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: OpListAll, Err: err}
	}
	resp, err := handleGetResponse(code, body, OpListAll)
	if err != nil {
		return nil, err
	}
	return []types.Media{*resp}, nil
}

func handleGetResponse(code int, body []byte, op string) (*types.Media, error) {
	switch code {
	case http.StatusOK:
		var resp otakotakuResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, &types.Error{Source: types.SourceOtakotaku, Op: op, Err: fmt.Errorf("parse: %w", err)}
		}
		m := resp.toMedia()
		return m, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: op, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: op, Err: types.ErrRateLimited}
	default:
		return nil, &types.Error{Source: types.SourceOtakotaku, Op: op, Err: fmt.Errorf("status %d", code)}
	}
}
