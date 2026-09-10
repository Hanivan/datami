package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

const (
	OpGet                    = "get"
	OpSearch                 = "search"
	OpGetByMalID             = "getbymalid"
	OpGetManga               = "getmanga"
	OpSearchManga            = "searchmanga"
	OpGetMangaByMalID        = "getmangabymalid"
	OpGetViewer              = "getviewer"
	OpGetMediaListEntry      = "getmedialistentry"
	OpGetMediaListCollection = "getmedialistcollection"
	OpSaveMediaListEntry     = "savemedialistentry"
	OpDeleteMediaListEntry   = "deletemedialistentry"
	OpGetCharacters          = "getcharacters"
	OpGetStaff               = "getstaff"
	OpGetTagsLinks           = "gettagslinks"
)

func (c *Client) Get(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGet, Err: types.ErrInvalidInput}
	}
	query := `query ($id: Int) { Media(id: $id) { id type title { romaji english native } description startDate { year month day } endDate { year month day } status format episodes duration chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { nodes { id name } } coverImage { medium large } bannerImage relations { edges { relationType node { id title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings stats } }`
	vars := map[string]any{"id": parseID(id)}
	return c.doGraphQL(ctx, query, vars, OpGet)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearch, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	gql := `query ($search: String, $perPage: Int) { Page(perPage: $perPage) { Media(search: $search) { id type title { romaji english native } description startDate { year month day } endDate { year month day } status format episodes duration averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { nodes { id name } } coverImage { medium large } bannerImage synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings stats } } }`
	vars := map[string]any{"search": query, "perPage": limit}
	return c.doSearchGraphQL(ctx, gql, vars)
}

func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if malID == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetByMalID, Err: types.ErrInvalidInput}
	}
	query := `query ($malId: Int) { Media(idMal: $malId) { id type title { romaji english native } description startDate { year month day } endDate { year month day } status format episodes duration chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { nodes { id name } } coverImage { medium large } bannerImage relations { edges { relationType node { id title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings stats } }`
	vars := map[string]any{"malId": parseID(malID)}
	return c.doGraphQL(ctx, query, vars, OpGetByMalID)
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetManga, Err: types.ErrInvalidInput}
	}
	query := `query ($id: Int) { Media(id: $id, type: MANGA) { id type title { romaji english native } description startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage relations { edges { relationType node { id title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings stats } }`
	vars := map[string]any{"id": parseID(id)}
	return c.doGraphQL(ctx, query, vars, OpGetManga)
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearchManga, Err: types.ErrInvalidInput}
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	gql := `query ($search: String, $perPage: Int) { Page(perPage: $perPage) { Media(search: $search, type: MANGA) { id type title { romaji english native } description startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings stats } } }`
	vars := map[string]any{"search": query, "perPage": limit}
	return c.doSearchGraphQL(ctx, gql, vars)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if malID == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetMangaByMalID, Err: types.ErrInvalidInput}
	}
	query := `query ($malId: Int) { Media(idMal: $malId, type: MANGA) { id type title { romaji english native } description startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage relations { edges { relationType node { id title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings stats } }`
	vars := map[string]any{"malId": parseID(malID)}
	return c.doGraphQL(ctx, query, vars, OpGetMangaByMalID)
}

func (c *Client) doGraphQL(ctx context.Context, query string, vars map[string]any, op string) (*types.Media, error) {
	payload := map[string]any{"query": query, "variables": vars}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("marshal: %w", err)}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("request: %w", err)}
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeJSON)
	req.Header.Set(httpx.HeaderAccept, httpx.ContentTypeJSON)
	if c.token != "" {
		req.Header.Set(httpx.HeaderAuthorization, httpx.BearerPrefix+c.token)
	}
	code, respBody, err := c.client.Do(req)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: err}
	}
	if code != http.StatusOK {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("status %d", code)}
	}
	var result mediaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("decode: %w", err)}
	}
	if result.Data.Media == nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: types.ErrNotFound}
	}
	return result.Data.Media.toMedia(), nil
}

func (c *Client) doSearchGraphQL(ctx context.Context, query string, vars map[string]any) ([]types.Media, error) {
	payload := map[string]any{"query": query, "variables": vars}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearch, Err: fmt.Errorf("marshal: %w", err)}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearch, Err: fmt.Errorf("request: %w", err)}
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeJSON)
	req.Header.Set(httpx.HeaderAccept, httpx.ContentTypeJSON)
	if c.token != "" {
		req.Header.Set(httpx.HeaderAuthorization, httpx.BearerPrefix+c.token)
	}
	code, respBody, err := c.client.Do(req)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearch, Err: err}
	}
	if code != http.StatusOK {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearch, Err: fmt.Errorf("status %d", code)}
	}
	var result searchResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearch, Err: fmt.Errorf("decode: %w", err)}
	}
	results := make([]types.Media, 0, len(result.Data.Page.Media))
	for _, m := range result.Data.Page.Media {
		results = append(results, *m.toMedia())
	}
	return results, nil
}

// Non-goal methods (not used in mono-crawler catalog path). Stubbed with ErrNotImplemented.
func (c *Client) GetViewer(ctx context.Context) (*types.UserInfo, error) {
	return nil, &types.Error{Source: types.SourceAniList, Op: OpGetViewer, Err: types.ErrNotImplemented}
}

func (c *Client) GetMediaListEntry(ctx context.Context, mediaID, userID int) (*types.MediaListEntry, error) {
	return nil, &types.Error{Source: types.SourceAniList, Op: OpGetMediaListEntry, Err: types.ErrNotImplemented}
}

func (c *Client) GetMediaListCollection(ctx context.Context, rt string, userID int) (*types.MediaListCollection, error) {
	return nil, &types.Error{Source: types.SourceAniList, Op: OpGetMediaListCollection, Err: types.ErrNotImplemented}
}

func (c *Client) SaveMediaListEntry(ctx context.Context, entry types.MediaListEntry) error {
	return &types.Error{Source: types.SourceAniList, Op: OpSaveMediaListEntry, Err: types.ErrNotImplemented}
}

func (c *Client) DeleteMediaListEntry(ctx context.Context, listEntryID int) error {
	return &types.Error{Source: types.SourceAniList, Op: OpDeleteMediaListEntry, Err: types.ErrNotImplemented}
}

// GetCharacters returns the character edges of a media by AniList id,
// walking pages (AniList caps at 25 per page).
func (c *Client) GetCharacters(ctx context.Context, id string) ([]types.Character, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetCharacters, Err: types.ErrInvalidInput}
	}
	var all []characterEdge
	vars := map[string]any{"id": parseID(id)}
	for page := 1; page <= 10; page++ {
		pageVars := copyVars(vars)
		pageVars["page"] = page
		pageVars["perPage"] = 25
		query := `query ($id: Int, $page: Int, $perPage: Int) { Media(id: $id) { characters(page: $page, perPage: $perPage, sort: ROLE) { pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } description } voiceActors { id name { first middle last full native alternative } image { medium large } languageV2 } } } } } }`
		conn, err := c.doCharConn(ctx, query, pageVars, OpGetCharacters)
		if err != nil {
			return nil, err
		}
		if conn == nil {
			break
		}
		all = append(all, conn.Edges...)
		if !conn.PageInfo.HasNextPage {
			break
		}
	}
	out := make([]types.Character, 0, len(all))
	for _, e := range all {
		out = append(out, *e.toCharacter())
	}
	return out, nil
}

// GetStaff returns the staff edges of a media by AniList id.
func (c *Client) GetStaff(ctx context.Context, id string) ([]types.Person, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetStaff, Err: types.ErrInvalidInput}
	}
	var all []staffEdge
	vars := map[string]any{"id": parseID(id)}
	for page := 1; page <= 10; page++ {
		pageVars := copyVars(vars)
		pageVars["page"] = page
		pageVars["perPage"] = 25
		query := `query ($id: Int, $page: Int, $perPage: Int) { Media(id: $id) { staff(page: $page, perPage: $perPage, sort: RELEVANCE) { pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } primaryOccupations } } } } }`
		conn, err := c.doStaffConn(ctx, query, pageVars, OpGetStaff)
		if err != nil {
			return nil, err
		}
		if conn == nil {
			break
		}
		all = append(all, conn.Edges...)
		if !conn.PageInfo.HasNextPage {
			break
		}
	}
	out := make([]types.Person, 0, len(all))
	for _, e := range all {
		out = append(out, *e.toPerson())
	}
	return out, nil
}

// TagsLinks holds a media's tags + external links from one fetch.
type TagsLinks struct {
	Tags          []Tag
	ExternalLinks []ExternalLink
}

type Tag struct {
	Name             string
	Category         string
	Rank             *int
	IsGeneralSpoiler bool
	IsMediaSpoiler   bool
	IsAdult          bool
}

type ExternalLink struct {
	ID         int
	URL        string
	Site       string
	SiteID     *int
	Type       string
	Language   string
	Notes      string
	IsDisabled bool
}

// GetTagsAndLinks returns a media's tags and external links by AniList id.
func (c *Client) GetTagsAndLinks(ctx context.Context, id string) (*TagsLinks, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetTagsLinks, Err: types.ErrInvalidInput}
	}
	query := `query ($id: Int) { Media(id: $id) { tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } externalLinks { id url site siteId type language notes isDisabled } } }`
	vars := map[string]any{"id": parseID(id)}
	body, err := c.post(ctx, query, vars, OpGetTagsLinks)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data struct {
			Media *struct {
				Tags          []tag          `json:"tags"`
				ExternalLinks []externalLink `json:"externalLinks"`
			} `json:"Media"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetTagsLinks, Err: fmt.Errorf("decode: %w", err)}
	}
	if result.Data.Media == nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetTagsLinks, Err: types.ErrNotFound}
	}
	tl := &TagsLinks{}
	for _, t := range result.Data.Media.Tags {
		tl.Tags = append(tl.Tags, Tag{
			Name: t.Name, Category: t.Category, Rank: t.Rank,
			IsGeneralSpoiler: t.IsGeneralSpoiler, IsMediaSpoiler: t.IsMediaSpoiler, IsAdult: t.IsAdult,
		})
	}
	for _, l := range result.Data.Media.ExternalLinks {
		tl.ExternalLinks = append(tl.ExternalLinks, ExternalLink{
			ID: l.ID, URL: l.URL, Site: l.Site, SiteID: l.SiteID, Type: l.Type,
			Language: l.Language, Notes: l.Notes, IsDisabled: l.IsDisabled,
		})
	}
	return tl, nil
}

func (c *Client) doCharConn(ctx context.Context, query string, vars map[string]any, op string) (*charConn, error) {
	body, err := c.post(ctx, query, vars, op)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data struct {
			Media *struct {
				Characters *charConn `json:"characters"`
			} `json:"Media"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("decode: %w", err)}
	}
	if result.Data.Media == nil {
		return nil, nil
	}
	return result.Data.Media.Characters, nil
}

func (c *Client) doStaffConn(ctx context.Context, query string, vars map[string]any, op string) (*staffConn, error) {
	body, err := c.post(ctx, query, vars, op)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data struct {
			Media *struct {
				Staff *staffConn `json:"staff"`
			} `json:"Media"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("decode: %w", err)}
	}
	if result.Data.Media == nil {
		return nil, nil
	}
	return result.Data.Media.Staff, nil
}

// post marshals + sends a GraphQL query, returns the raw response body.
func (c *Client) post(ctx context.Context, query string, vars map[string]any, op string) ([]byte, error) {
	payload := map[string]any{"query": query, "variables": vars}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("marshal: %w", err)}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("request: %w", err)}
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeJSON)
	req.Header.Set(httpx.HeaderAccept, httpx.ContentTypeJSON)
	if c.token != "" {
		req.Header.Set(httpx.HeaderAuthorization, httpx.BearerPrefix+c.token)
	}
	code, respBody, err := c.client.Do(req)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: err}
	}
	if code != http.StatusOK {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("status %d", code)}
	}
	return respBody, nil
}

func copyVars(vars map[string]any) map[string]any {
	out := make(map[string]any, len(vars)+2)
	for k, v := range vars {
		out[k] = v
	}
	return out
}

func parseID(s string) int {
	id, _ := strconv.Atoi(s)
	return id
}
