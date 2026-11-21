package anilist

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Hanivan/datami/types"
)

// maxBatch is AniList's perPage ceiling.
const maxBatch = 50

// GetByMalIDs fetches the ANIME whose MAL ids are malIDs in one request (plus
// packed follow-ups for works with more cast/staff than the embedded pages hold). The
// result is keyed by MAL id; ids AniList does not know are absent. At most 50
// ids; empty input returns an empty map without a request.
func (c *Client) GetByMalIDs(ctx context.Context, malIDs []string, opts ...types.MediaOpts) (map[string]*types.Media, error) {
	return c.getByMalIDs(ctx, malIDs, "ANIME", c.animeSel(), OpGetByMalIDs, opts)
}

// GetMangaByMalIDs is GetByMalIDs for the manga namespace.
func (c *Client) GetMangaByMalIDs(ctx context.Context, malIDs []string, opts ...types.MediaOpts) (map[string]*types.Media, error) {
	return c.getByMalIDs(ctx, malIDs, "MANGA", c.mangaSel(), OpGetMangaByMalIDs, opts)
}

func (c *Client) getByMalIDs(ctx context.Context, malIDs []string, typ, sel, op string, opts []types.MediaOpts) (map[string]*types.Media, error) {
	out := map[string]*types.Media{}
	if len(malIDs) == 0 {
		return out, nil
	}
	fail := func(err error) (map[string]*types.Media, error) {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: err}
	}
	if len(malIDs) > maxBatch {
		return fail(types.ErrInvalidInput)
	}
	ids := make([]int, len(malIDs))
	for i, s := range malIDs {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fail(types.ErrInvalidInput)
		}
		ids[i] = n
	}
	var mo types.MediaOpts
	if len(opts) > 0 {
		mo = opts[0]
	}
	// embedPages expects the query to end with the media close and the query
	// close; the Page wrapper adds one more close, so it is appended after.
	q := `query ($ids: [Int], $type: MediaType, $perPage: Int) { Page(page: 1, perPage: $perPage) { media(idMal_in: $ids, type: $type) { ` + sel + ` } } }`
	q = c.embedPages(q[:len(q)-2], mo) + " }"
	vars := map[string]any{"ids": ids, "type": typ, "perPage": len(ids)}
	body, err := c.post(ctx, q, vars, op)
	if err != nil {
		return nil, err
	}
	var typed struct {
		Data struct {
			Page struct {
				Media []*media `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}
	var rawResp struct {
		Data struct {
			Page struct {
				Media []map[string]json.RawMessage `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &typed); err != nil {
		return fail(fmt.Errorf("decode: %w", err))
	}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return fail(fmt.Errorf("decode: %w", err))
	}
	var pend []pending
	for i, md := range typed.Data.Page.Media {
		if md == nil || md.MALID == nil {
			continue
		}
		m := md.toMedia()
		if mo.CharacterPages > 0 || mo.StaffPages > 0 {
			p, err := mergeEmbeddedMap(rawResp.Data.Page.Media[i], md.ID, mo, m)
			if err != nil {
				return nil, err
			}
			pend = append(pend, p...)
		}
		out[strconv.Itoa(*md.MALID)] = m
	}
	if err := c.followUps(ctx, pend, op); err != nil {
		return nil, err
	}
	return out, nil
}
