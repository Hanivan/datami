package mal

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
		return nil, helpers.Wrap(types.SourceMAL, OpGet, types.ErrInvalidInput)
	}
	u := fmt.Sprintf("%s/anime/%s?fields=id,title,main_picture,synopsis,start_date,end_date,status,media_type,num_episodes,average_episode_duration,mean,rank,popularity,num_list_users,num_scoring_users,num_favorites,nsfw,genres,studios,alternative_titles,related_anime{media_type},related_manga{media_type},start_season,broadcast,source,rating,pictures,background,recommendations,statistics,my_list_status,created_at,updated_at", c.baseURL, url.PathEscape(id))
	code, body, retryAfter, err := c.client.Get(ctx, OpGet, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGet, err)
	}
	return handleResponse(code, body, OpGet, retryAfter)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, helpers.Wrap(types.SourceMAL, OpSearch, types.ErrInvalidInput)
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime?q=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(query), limit, animeListFields)
	code, body, retryAfter, err := c.client.Get(ctx, OpSearch, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpSearch, err)
	}
	return handleSearchResponse(code, body, retryAfter)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return c.Get(ctx, malID)
}
