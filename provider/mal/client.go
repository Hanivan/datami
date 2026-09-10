package mal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGet, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/anime/%s?fields=id,title,main_picture,synopsis,start_date,end_date,status,media_type,num_episodes,average_episode_duration,mean,rank,popularity,num_list_users,num_scoring_users,num_favorites,nsfw,genres,studios,alternative_titles,related_anime{media_type},related_manga{media_type},start_season,broadcast,source,rating,pictures,background,recommendations,statistics,my_list_status,created_at,updated_at", c.baseURL, url.PathEscape(id))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGet, Err: err}
	}
	return handleResponse(code, body, "get")
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime?q=%s&limit=%d", c.baseURL, url.QueryEscape(query), limit)
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
	return handleMangaResponse(code, body, "getmanga")
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/manga?q=%s&limit=%d", c.baseURL, url.QueryEscape(query), limit)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: err}
	}
	return handleMangaSearchResponse(code, body)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	return c.GetManga(ctx, malID)
}

func (c *Client) GetSeasonal(ctx context.Context, year int, season string, opts types.SearchOpts) ([]types.Media, error) {
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	u := fmt.Sprintf("%s/anime/season/%d/%s?limit=%d", c.baseURL, year, url.PathEscape(season), limit)
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
	u := fmt.Sprintf("%s/anime/ranking?ranking_type=%s&limit=%d", c.baseURL, url.QueryEscape(rankingType), limit)
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
	u := fmt.Sprintf("%s/manga/ranking?ranking_type=%s&limit=%d", c.baseURL, url.QueryEscape(rankingType), limit)
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetMangaRanking, Err: err}
	}
	return handleMangaRankingResponse(code, body)
}

func (c *Client) GetAnimeCharacters(ctx context.Context, animeID string) ([]types.Character, error) {
	if animeID == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/anime/%s/characters?limit=100&fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography", c.baseURL, url.PathEscape(animeID))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: err}
	}
	return handleCharactersResponse(code, body)
}

func (c *Client) GetCharacter(ctx context.Context, charID string) (*types.Character, error) {
	if charID == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/characters/%s?fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography,animeography", c.baseURL, url.PathEscape(charID))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: err}
	}
	return handleCharacterResponse(code, body)
}

func (c *Client) GetPerson(ctx context.Context, personID string) (*types.Person, error) {
	if personID == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/people/%s?fields=first_name,last_name,main_picture,birthday,num_favorites,more", c.baseURL, url.PathEscape(personID))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: err}
	}
	return handlePersonResponse(code, body)
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
	switch code {
	case http.StatusOK:
		var a anime
		if err := json.Unmarshal(body, &a); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: fmt.Errorf("parse: %w", err)}
		}
		return a.toMedia(), nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: fmt.Errorf("status %d", code)}
	}
}

func handleSearchResponse(code int, body []byte) ([]types.Media, error) {
	switch code {
	case http.StatusOK:
		var resp searchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: fmt.Errorf("parse: %w", err)}
		}
		results := make([]types.Media, 0, len(resp.Data))
		for _, node := range resp.Data {
			m := node.Node.toMedia()
			results = append(results, *m)
		}
		return results, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearch, Err: fmt.Errorf("status %d", code)}
	}
}

func handleMangaResponse(code int, body []byte, op string) (*types.Media, error) {
	switch code {
	case http.StatusOK:
		var m manga
		if err := json.Unmarshal(body, &m); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: fmt.Errorf("parse: %w", err)}
		}
		return m.toMedia(), nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: op, Err: fmt.Errorf("status %d", code)}
	}
}

func handleMangaSearchResponse(code int, body []byte) ([]types.Media, error) {
	switch code {
	case http.StatusOK:
		var resp mangaSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: fmt.Errorf("parse: %w", err)}
		}
		results := make([]types.Media, 0, len(resp.Data))
		for _, node := range resp.Data {
			results = append(results, *node.Node.toMedia())
		}
		return results, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpSearchManga, Err: fmt.Errorf("status %d", code)}
	}
}

func handleRankingResponse(code int, body []byte) ([]types.Media, error) {
	switch code {
	case http.StatusOK:
		var resp rankingResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: OpGetRanking, Err: fmt.Errorf("parse: %w", err)}
		}
		results := make([]types.Media, 0, len(resp.Data))
		for _, node := range resp.Data {
			m := node.Node.toMedia()
			results = append(results, *m)
		}
		return results, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetRanking, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetRanking, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetRanking, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetRanking, Err: fmt.Errorf("status %d", code)}
	}
}

func handleMangaRankingResponse(code int, body []byte) ([]types.Media, error) {
	switch code {
	case http.StatusOK:
		var resp mangaSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: OpGetMangaRanking, Err: fmt.Errorf("parse: %w", err)}
		}
		results := make([]types.Media, 0, len(resp.Data))
		for _, node := range resp.Data {
			results = append(results, *node.Node.toMedia())
		}
		return results, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetMangaRanking, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetMangaRanking, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetMangaRanking, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetMangaRanking, Err: fmt.Errorf("status %d", code)}
	}
}

func handleCharactersResponse(code int, body []byte) ([]types.Character, error) {
	switch code {
	case http.StatusOK:
		var resp charactersResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: fmt.Errorf("parse: %w", err)}
		}
		results := make([]types.Character, 0, len(resp.Data))
		for _, ac := range resp.Data {
			c := ac.toCharacter()
			results = append(results, *c)
		}
		return results, nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: fmt.Errorf("status %d", code)}
	}
}

func handleCharacterResponse(code int, body []byte) (*types.Character, error) {
	switch code {
	case http.StatusOK:
		var c character
		if err := json.Unmarshal(body, &c); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: fmt.Errorf("parse: %w", err)}
		}
		return c.toCharacter(), nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: fmt.Errorf("status %d", code)}
	}
}

func handlePersonResponse(code int, body []byte) (*types.Person, error) {
	switch code {
	case http.StatusOK:
		var p person
		if err := json.Unmarshal(body, &p); err != nil {
			return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: fmt.Errorf("parse: %w", err)}
		}
		return p.toPerson(), nil
	case http.StatusNotFound:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: types.ErrNotFound}
	case http.StatusTooManyRequests:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: types.ErrRateLimited}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: types.ErrAuthFailed}
	default:
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: fmt.Errorf("status %d", code)}
	}
}
