package anilist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hanivan/datami/types"
)

// GetStaff returns the staff edges of a media by AniList id.
func (c *Client) GetStaff(ctx context.Context, id string) ([]types.Person, error) {
	return c.staffBy(ctx, "id", id, "")
}

// GetStaffByMalID returns the staff edges of a media addressed by its MAL id.
// See GetCharactersByMalID on why the id namespace matters.
func (c *Client) GetStaffByMalID(ctx context.Context, malID, mediaType string) ([]types.Person, error) {
	return c.staffBy(ctx, "idMal", malID, mediaType)
}

func (c *Client) staffBy(ctx context.Context, idField, id, mediaType string) ([]types.Person, error) {
	if id == "" {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetStaff, Err: types.ErrInvalidInput}
	}
	sel := `staff(page: $page, perPage: $perPage, sort: RELEVANCE) { pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } primaryOccupations } } }`
	query, vars := mediaPageQuery(idField, id, mediaType, sel)
	var all []staffEdge
	for page := 1; page <= 10; page++ {
		pageVars := copyVars(vars)
		pageVars["page"] = page
		pageVars["perPage"] = 25
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
