package mal

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

// GetSeasonal returns one page of a season. opts.Page is 1-based and maps to
// MAL's `offset`; MAL reports no total, so the caller detects the last page by
// receiving fewer than opts.Limit items.
func (c *Client) GetSeasonal(ctx context.Context, year int, season string, opts types.SearchOpts) ([]types.Media, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	page := opts.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	u := fmt.Sprintf("%s/anime/season/%d/%s?limit=%d&offset=%d&fields=%s", c.baseURL, year, url.PathEscape(season), limit, offset, animeListFields)
	code, body, retryAfter, err := c.client.Get(ctx, OpGetSeasonal, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGetSeasonal, err)
	}
	return handleSearchResponse(code, body, retryAfter)
}

func (c *Client) GetRanking(ctx context.Context, rankingType string, opts types.SearchOpts) ([]types.Media, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime/ranking?ranking_type=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(rankingType), limit, animeListFields)
	code, body, retryAfter, err := c.client.Get(ctx, OpGetRanking, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGetRanking, err)
	}
	return handleRankingResponse(code, body, retryAfter)
}

func (c *Client) GetMangaRanking(ctx context.Context, rankingType string, opts types.SearchOpts) ([]types.Media, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/manga/ranking?ranking_type=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(rankingType), limit, mangaListFields)
	code, body, retryAfter, err := c.client.Get(ctx, OpGetMangaRanking, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGetMangaRanking, err)
	}
	return handleMangaListResponse(code, body, OpGetMangaRanking, retryAfter)
}
