package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

func (c *Client) doGraphQL(ctx context.Context, query string, vars map[string]any, op string, opts []types.MediaOpts) (*types.Media, error) {
	var mo types.MediaOpts
	if len(opts) > 0 {
		mo = opts[0]
	}
	query = c.embedPages(query, mo)
	body, err := c.post(ctx, query, vars, op)
	if err != nil {
		return nil, err
	}
	var result mediaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, helpers.Wrap(types.SourceAniList, op, fmt.Errorf("decode: %w", err))
	}
	if result.Data.Media == nil {
		return nil, helpers.Wrap(types.SourceAniList, op, types.ErrNotFound)
	}
	m := result.Data.Media.toMedia()
	if mo.CharacterPages > 0 || mo.StaffPages > 0 {
		pend, err := mergeEmbedded(body, result.Data.Media.ID, mo, m, op)
		if err != nil {
			return nil, err
		}
		if err := c.followUps(ctx, pend); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (c *Client) doSearchGraphQL(ctx context.Context, query string, vars map[string]any) ([]types.Media, error) {
	body, err := c.post(ctx, query, vars, OpSearch)
	if err != nil {
		return nil, err
	}
	var result searchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, helpers.Wrap(types.SourceAniList, OpSearch, fmt.Errorf("decode: %w", err))
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
		return nil, helpers.Wrap(types.SourceAniList, op, fmt.Errorf("marshal: %w", err))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, helpers.Wrap(types.SourceAniList, op, fmt.Errorf("request: %w", err))
	}
	req.Header.Set(httpx.HeaderContentType, httpx.ContentTypeJSON)
	req.Header.Set(httpx.HeaderAccept, httpx.ContentTypeJSON)
	if c.token != "" {
		req.Header.Set(httpx.HeaderAuthorization, httpx.BearerPrefix+c.token)
	}
	code, respBody, retryAfter, err := c.client.Do(op, req)
	if err != nil {
		return nil, helpers.Wrap(types.SourceAniList, op, err)
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
