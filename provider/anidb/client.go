package anidb

import (
	"context"
	"fmt"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

// Compile-time check that Client satisfies types.Provider.
var _ types.Provider = (*Client)(nil)

func wrap(op string, err error) error {
	return helpers.Wrap(types.SourceAniDB, op, err)
}

// guard returns ErrInvalidInput for an Unconfigured client.
func (c *Client) guard(op string) error {
	if c.unconfigured {
		return wrap(op, fmt.Errorf("%w: anidb client name/version not configured", types.ErrInvalidInput))
	}
	return nil
}

func (c *Client) Get(ctx context.Context, aid string) (*types.Media, error) {
	a, err := c.fetchAnime(ctx, OpGet, aid)
	if err != nil {
		return nil, err
	}
	return a.toMedia(c.now()), nil
}

// Search matches query against the cached anime-titles dump (see titles.go).
// Results carry only titles and the AniDB ExternalID; call Get for detail.
func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if err := c.guard(OpSearch); err != nil {
		return nil, err
	}
	q := normalizeTitle(query)
	if q == "" {
		return nil, wrap(OpSearch, types.ErrInvalidInput)
	}
	idx, err := c.titleIndex(ctx)
	if err != nil {
		return nil, err
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = httpx.DefaultSearchLimit
	}
	return searchIndex(idx, q, limit), nil
}

func (c *Client) GetCharacters(ctx context.Context, aid string) ([]types.Character, error) {
	a, err := c.fetchAnime(ctx, OpGetCharacters, aid)
	if err != nil {
		return nil, err
	}
	return a.toCharacters(aid), nil
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if err := c.guard(OpGetByMalID); err != nil {
		return nil, err
	}
	return nil, wrap(OpGetByMalID, types.ErrNotImplemented)
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if err := c.guard(OpGetManga); err != nil {
		return nil, err
	}
	return nil, wrap(OpGetManga, types.ErrNotImplemented)
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if err := c.guard(OpSearchManga); err != nil {
		return nil, err
	}
	return nil, wrap(OpSearchManga, types.ErrNotImplemented)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if err := c.guard(OpGetMangaByMalID); err != nil {
		return nil, err
	}
	return nil, wrap(OpGetMangaByMalID, types.ErrNotImplemented)
}
