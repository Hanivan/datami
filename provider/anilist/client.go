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
	query := `query ($id: Int) { Media(id: $id) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format episodes duration chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { edges { isMain node { id name isAnimationStudio } } } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } }`
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
	gql := `query ($search: String, $perPage: Int) { Page(perPage: $perPage) { Media(search: $search) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format episodes duration averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { edges { isMain node { id name isAnimationStudio } } } coverImage { medium large } bannerImage synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } } }`
	vars := map[string]any{"search": query, "perPage": limit}
	return c.doSearchGraphQL(ctx, gql, vars)
}

// GetByMalID fetches the ANIME whose MAL id is malID. The type filter is not
// optional: MAL anime and manga ids are separate namespaces, so an unfiltered
// idMal lookup can resolve to the manga that happens to share the number. Use
// GetMangaByMalID for the manga namespace.
func (c *Client) GetByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if malID == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetByMalID, Err: types.ErrInvalidInput}
	}
	query := `query ($malId: Int, $type: MediaType) { Media(idMal: $malId, type: $type) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format episodes duration chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { edges { isMain node { id name isAnimationStudio } } } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } }`
	vars := map[string]any{"malId": parseID(malID), "type": "ANIME"}
	return c.doGraphQL(ctx, query, vars, OpGetByMalID)
}

func (c *Client) GetManga(ctx context.Context, id string) (*types.Media, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetManga, Err: types.ErrInvalidInput}
	}
	query := `query ($id: Int) { Media(id: $id, type: MANGA) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } }`
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
	gql := `query ($search: String, $perPage: Int) { Page(perPage: $perPage) { Media(search: $search, type: MANGA) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } } }`
	vars := map[string]any{"search": query, "perPage": limit}
	return c.doSearchGraphQL(ctx, gql, vars)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string) (*types.Media, error) {
	if malID == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetMangaByMalID, Err: types.ErrInvalidInput}
	}
	query := `query ($malId: Int) { Media(idMal: $malId, type: MANGA) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } }`
	vars := map[string]any{"malId": parseID(malID)}
	return c.doGraphQL(ctx, query, vars, OpGetMangaByMalID)
}

func (c *Client) doGraphQL(ctx context.Context, query string, vars map[string]any, op string) (*types.Media, error) {
	body, err := c.post(ctx, query, vars, op)
	if err != nil {
		return nil, err
	}
	var result mediaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("decode: %w", err)}
	}
	if result.Data.Media == nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: types.ErrNotFound}
	}
	return result.Data.Media.toMedia(), nil
}

func (c *Client) doSearchGraphQL(ctx context.Context, query string, vars map[string]any) ([]types.Media, error) {
	body, err := c.post(ctx, query, vars, OpSearch)
	if err != nil {
		return nil, err
	}
	var result searchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpSearch, Err: fmt.Errorf("decode: %w", err)}
	}
	results := make([]types.Media, 0, len(result.Data.Page.Media))
	for _, m := range result.Data.Page.Media {
		results = append(results, *m.toMedia())
	}
	return results, nil
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
	code, respBody, retryAfter, err := c.client.Do(op, req)
	if err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: err}
	}
	if code != http.StatusOK {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: statusErr(code), Status: code, RetryAfter: retryAfter}
	}
	return respBody, nil
}

// statusErr maps an HTTP status onto the types sentinel that describes it, so
// callers can errors.Is on the condition instead of scraping the message. A
// status with no sentinel keeps the numeric text.
//
// 403 is rate-limiting here, not authorization: AniList answers 403 to every
// query during a global suspension or a per-IP block, and in both cases the
// remedy is to stop calling for a while, exactly as for 429.
func statusErr(code int) error {
	switch code {
	case http.StatusTooManyRequests, http.StatusForbidden:
		return types.ErrRateLimited
	case http.StatusNotFound:
		return types.ErrNotFound
	case http.StatusUnauthorized:
		return types.ErrAuthFailed
	case http.StatusBadRequest:
		return types.ErrInvalidInput
	}
	return fmt.Errorf("status %d", code)
}

func copyVars(vars map[string]any) map[string]any {
	out := make(map[string]any, len(vars)+2)
	for k, v := range vars {
		out[k] = v
	}
	return out
}

// mediaPageQuery builds a paginated `Media(<idField>: $<idField>[, type: $type])`
// query around sel, plus its variables. idField is "id" (AniList) or "idMal";
// mediaType "" omits the type filter.
func mediaPageQuery(idField, id, mediaType, sel string) (string, map[string]any) {
	vars := map[string]any{idField: parseID(id)}
	args := "$" + idField + ": Int, $page: Int, $perPage: Int"
	mediaArgs := idField + ": $" + idField
	if mediaType != "" {
		args += ", $type: MediaType"
		mediaArgs += ", type: $type"
		vars["type"] = mediaType
	}
	return "query (" + args + ") { Media(" + mediaArgs + ") { " + sel + " } }", vars
}

func parseID(s string) int {
	id, _ := strconv.Atoi(s)
	return id
}

// desc is the description field selection. nil asHTML means HTML, so Media
// and Character descriptions share one format by default.
func (c *Client) desc() string {
	return "description(asHtml: " + strconv.FormatBool(c.asHTML == nil || *c.asHTML) + ")"
}
