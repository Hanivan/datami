package mal

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

const (
	OpGet                   = "get"
	OpSearch                = "search"
	OpGetManga              = "getmanga"
	OpSearchManga           = "searchmanga"
	OpGetSeasonal           = "getseasonal"
	OpGetRanking            = "getranking"
	OpGetMangaRanking       = "getmangaranking"
	OpGetAnimeCharacters    = "getanimecharacters"
	OpGetCharacter          = "getcharacter"
	OpGetPerson             = "getperson"
	OpGetForumBoards        = "getforumboards"
	OpGetForumTopic         = "getforumtopic"
	OpSearchForumTopics     = "searchforumtopics"
	OpGetUserAnimeList      = "getuseranimelist"
	OpGetUserMangaList      = "getusermangalist"
	OpGetUserInfo           = "getuserinfo"
	OpUpdateAnimeListStatus = "updateanimeliststatus"
	OpDeleteAnimeListItem   = "deleteanimelistitem"
)

// MAL's list endpoints (search, seasonal, ranking) return only id, title and
// main_picture unless `fields` asks for more — unlike the by-id endpoints,
// which fail loudly on a bad field. A list request for an unsupported field is
// silently dropped from the response, so anime and manga need separate sets:
// asking a manga endpoint for num_episodes yields no error and no value.
const (
	animeListFields = "id,title,main_picture,media_type,status,num_episodes,mean"
	mangaListFields = "id,title,main_picture,media_type,status,num_chapters,num_volumes,mean"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGet, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/anime/%s?fields=id,title,main_picture,synopsis,start_date,end_date,status,media_type,num_episodes,average_episode_duration,mean,rank,popularity,num_list_users,num_scoring_users,num_favorites,nsfw,genres,studios,alternative_titles,related_anime{media_type},related_manga{media_type},start_season,broadcast,source,rating,pictures,background,recommendations,statistics,my_list_status,created_at,updated_at", c.baseURL, url.PathEscape(id))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGet, Err: err}
	}
	return handleResponse(code, body, OpGet)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime?q=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(query), limit, animeListFields)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: err}
	}
	return handleSearchResponse(code, body)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return c.Get(ctx, malID)
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetManga, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/manga/%s?fields=id,title,main_picture,synopsis,start_date,end_date,status,media_type,num_chapters,num_volumes,mean,rank,popularity,num_list_users,num_scoring_users,num_favorites,nsfw,genres,alternative_titles,related_manga{media_type},related_anime{media_type},authors,serialization,pictures,background,recommendations,statistics,my_list_status,created_at,updated_at", c.baseURL, url.PathEscape(id))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetManga, Err: err}
	}
	return handleMangaResponse(code, body, OpGetManga)
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/manga?q=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(query), limit, mangaListFields)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: err}
	}
	return handleMangaListResponse(code, body, OpSearchManga)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return c.GetManga(ctx, malID)
}

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
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetSeasonal, Err: err}
	}
	return handleSearchResponse(code, body)
}

func (c *Client) GetRanking(ctx context.Context, rankingType string, opts types.SearchOpts) ([]types.Media, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime/ranking?ranking_type=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(rankingType), limit, animeListFields)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetRanking, Err: err}
	}
	return handleRankingResponse(code, body)
}

func (c *Client) GetMangaRanking(ctx context.Context, rankingType string, opts types.SearchOpts) ([]types.Media, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/manga/ranking?ranking_type=%s&limit=%d&fields=%s", c.baseURL, url.QueryEscape(rankingType), limit, mangaListFields)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetMangaRanking, Err: err}
	}
	return handleMangaListResponse(code, body, OpGetMangaRanking)
}

// Non-goal methods (not used in mono-crawler catalog path). Stubbed with ErrNotImplemented.
func (c *Client) GetForumBoards(ctx context.Context) ([]types.ForumBoard, error) {
	return nil, &types.Error{Source: types.SourceMAL, Op: OpGetForumBoards, Err: types.ErrNotImplemented}
}

func (c *Client) GetForumTopic(ctx context.Context, topicID string, limit int) ([]types.ForumPost, error) {
	return nil, &types.Error{Source: types.SourceMAL, Op: OpGetForumTopic, Err: types.ErrNotImplemented}
}

func (c *Client) SearchForumTopics(ctx context.Context, q string, limit int) ([]types.ForumTopic, error) {
	return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchForumTopics, Err: types.ErrNotImplemented}
}

func (c *Client) GetUserAnimeList(ctx context.Context, userName string, limit int) ([]types.Media, error) {
	return nil, &types.Error{Source: types.SourceMAL, Op: OpGetUserAnimeList, Err: types.ErrNotImplemented}
}

func (c *Client) GetUserMangaList(ctx context.Context, userName string, limit int) ([]types.Media, error) {
	return nil, &types.Error{Source: types.SourceMAL, Op: OpGetUserMangaList, Err: types.ErrNotImplemented}
}

func (c *Client) GetUserInfo(ctx context.Context) ([]types.UserInfo, error) {
	return nil, &types.Error{Source: types.SourceMAL, Op: OpGetUserInfo, Err: types.ErrNotImplemented}
}

func (c *Client) UpdateAnimeListStatus(ctx context.Context, animeID int, form url.Values) error {
	return &types.Error{Source: types.SourceMAL, Op: OpUpdateAnimeListStatus, Err: types.ErrNotImplemented}
}

func (c *Client) DeleteAnimeListItem(ctx context.Context, animeID int) error {
	return &types.Error{Source: types.SourceMAL, Op: OpDeleteAnimeListItem, Err: types.ErrNotImplemented}
}

func handleResponse(code int, body []byte, op string) (*types.Media, error) {
	a, err := helpers.DecodeJSON[anime](types.SourceMAL, op, code, body)
	if err != nil {
		return nil, err
	}
	return a.toMedia(), nil
}

func handleSearchResponse(code int, body []byte) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[searchResponse](types.SourceMAL, OpSearch, code, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, node := range resp.Data {
		results = append(results, *node.Node.toMedia())
	}
	return results, nil
}

func handleMangaResponse(code int, body []byte, op string) (*types.Media, error) {
	m, err := helpers.DecodeJSON[manga](types.SourceMAL, op, code, body)
	if err != nil {
		return nil, err
	}
	return m.toMedia(), nil
}

// handleMangaListResponse serves manga search and manga ranking (same shape).
func handleMangaListResponse(code int, body []byte, op string) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[mangaSearchResponse](types.SourceMAL, op, code, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, node := range resp.Data {
		results = append(results, *node.Node.toMedia())
	}
	return results, nil
}

func handleRankingResponse(code int, body []byte) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[rankingResponse](types.SourceMAL, OpGetRanking, code, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, node := range resp.Data {
		results = append(results, *node.Node.toMedia())
	}
	return results, nil
}
