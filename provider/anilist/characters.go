package anilist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hanivan/datami/types"
)

// GetCharacters returns the character edges of a media by AniList id,
// walking pages (AniList caps at 25 per page).
func (c *Client) GetCharacters(ctx context.Context, id string) ([]types.Character, error) {
	return c.charactersBy(ctx, "id", id, "")
}

// GetCharactersByMalID returns the character edges of a media addressed by its
// MAL id. AniList ids and MAL ids are SEPARATE namespaces, so passing a MAL id
// to GetCharacters resolves to an unrelated media (or none) - a caller holding
// a MAL id must use this. mediaType ("ANIME"/"MANGA") disambiguates, since one
// MAL id can exist in both namespaces; "" queries without the type filter.
func (c *Client) GetCharactersByMalID(ctx context.Context, malID, mediaType string) ([]types.Character, error) {
	return c.charactersBy(ctx, "idMal", malID, mediaType)
}

func (c *Client) charactersBy(ctx context.Context, idField, id, mediaType string) ([]types.Character, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetCharacters, Err: types.ErrInvalidInput}
	}
	sel := `characters(page: $page, perPage: $perPage, sort: ROLE) { pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } ` + c.desc() + ` } voiceActors { id name { first middle last full native alternative } image { medium large } languageV2 } } }`
	query, vars := mediaPageQuery(idField, id, mediaType, sel)
	var all []characterEdge
	for page := 1; page <= 10; page++ {
		pageVars := copyVars(vars)
		pageVars["page"] = page
		pageVars["perPage"] = 25
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
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("decode: %w", err)}
	}
	if result.Data.Media == nil {
		return nil, nil
	}
	return result.Data.Media.Characters, nil
}
