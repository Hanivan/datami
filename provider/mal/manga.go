package mal

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceMAL, OpGetManga, types.ErrInvalidInput)
	}
	u := fmt.Sprintf("%s/manga/%s?fields=id,title,main_picture,synopsis,start_date,end_date,status,media_type,num_chapters,num_volumes,mean,rank,popularity,num_list_users,num_scoring_users,num_favorites,nsfw,genres,alternative_titles,related_manga{media_type},related_anime{media_type},authors,serialization,pictures,background,recommendations,statistics,my_list_status,created_at,updated_at", c.baseURL, url.PathEscape(id))
	code, body, retryAfter, err := c.client.Get(ctx, OpGetManga, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGetManga, err)
	}
	return handleMangaResponse(code, body, OpGetManga, retryAfter)
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, helpers.Wrap(types.SourceMAL, OpSearchManga, types.ErrInvalidInput)
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/manga?q=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(query), limit, mangaListFields)
	code, body, retryAfter, err := c.client.Get(ctx, OpSearchManga, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpSearchManga, err)
	}
	return handleMangaListResponse(code, body, OpSearchManga, retryAfter)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return c.GetManga(ctx, malID)
}
