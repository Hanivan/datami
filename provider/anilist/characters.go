package anilist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// GetCharacters returns the character edges of a media by AniList id,
// walking pages (AniList caps at 25 per page).
func (c *Client) GetCharacters(ctx context.Context, id string, opts ...types.PageOpts) ([]types.Character, error) {
	return c.charactersBy(ctx, "id", id, "", opts...)
}

// GetCharactersByMalID returns the character edges of a media addressed by its
// MAL id. AniList ids and MAL ids are SEPARATE namespaces, so passing a MAL id
// to GetCharacters resolves to an unrelated media (or none) - a caller holding
// a MAL id must use this. mediaType ("ANIME"/"MANGA") disambiguates, since one
// MAL id can exist in both namespaces; "" queries without the type filter.
func (c *Client) GetCharactersByMalID(ctx context.Context, malID, mediaType string, opts ...types.PageOpts) ([]types.Character, error) {
	return c.charactersBy(ctx, "idMal", malID, mediaType, opts...)
}

func (c *Client) charactersBy(ctx context.Context, idField, id, mediaType string, opts ...types.PageOpts) ([]types.Character, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpGetCharacters, types.ErrInvalidInput)
	}
	sel := "characters(page: $page, perPage: $perPage, " + charSort + ") " + c.charBody()
	query, vars := mediaPageQuery(idField, id, mediaType, sel)
	var all []characterEdge
	start, max := resolvePages(opts)
	for page := start; page < start+max; page++ {
		pageVars := copyVars(vars)
		pageVars["page"] = page
		pageVars["perPage"] = pageSize
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
		return nil, helpers.Wrap(types.SourceAniList, op, fmt.Errorf("decode: %w", err))
	}
	if result.Data.Media == nil {
		return nil, nil
	}
	return result.Data.Media.Characters, nil
}
