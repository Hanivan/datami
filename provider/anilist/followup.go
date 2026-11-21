package anilist

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Hanivan/datami/types"
)

// followUpAliases caps the aliased connections packed into one request.
const followUpAliases = 8

// pending is one cast or staff connection with pages left to fetch: next is
// the page to ask for, last the connection's page cap.
type pending struct {
	media   *types.Media
	mediaID int
	staff   bool
	next    int
	last    int
}

// followUps fetches the remaining pages of every pending connection in
// rounds. Each round asks for the next page of every connection still
// pending, packed followUpAliases per request across works. A connection
// drops out when its page reports no next page, its Media answers null, or it
// reaches its cap. Any failed request fails the call.
func (c *Client) followUps(ctx context.Context, pend []pending, op string) error {
	for len(pend) > 0 {
		var keep []pending
		for start := 0; start < len(pend); start += followUpAliases {
			chunk := pend[start:min(start+followUpAliases, len(pend))]
			more, err := c.followUpRequest(ctx, chunk, op)
			if err != nil {
				return err
			}
			keep = append(keep, more...)
		}
		pend = keep
	}
	return nil
}

// followUpRequest sends one packed query for chunk, merges each alias's edges
// into its media, and returns the connections that continue.
func (c *Client) followUpRequest(ctx context.Context, chunk []pending, op string) ([]pending, error) {
	var b strings.Builder
	b.WriteString("query {")
	for i, p := range chunk {
		if p.staff {
			fmt.Fprintf(&b, " a%d: Media(id: %d) { staff(page: %d, perPage: %d, %s) %s }", i, p.mediaID, p.next, pageSize, staffSort, staffBody)
		} else {
			fmt.Fprintf(&b, " a%d: Media(id: %d) { characters(page: %d, perPage: %d, %s) %s }", i, p.mediaID, p.next, pageSize, charSort, c.charBody())
		}
	}
	b.WriteString(" }")
	body, err := c.post(ctx, b.String(), map[string]any{}, op)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data map[string]*struct {
			Characters *charConn  `json:"characters"`
			Staff      *staffConn `json:"staff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: op, Err: fmt.Errorf("decode: %w", err)}
	}
	var more []pending
	for i, p := range chunk {
		md := resp.Data[fmt.Sprintf("a%d", i)]
		if md == nil {
			continue
		}
		var next bool
		if p.staff {
			if md.Staff == nil {
				continue
			}
			for _, e := range md.Staff.Edges {
				p.media.Staff = append(p.media.Staff, *e.toPerson())
			}
			next = md.Staff.PageInfo.HasNextPage
		} else {
			if md.Characters == nil {
				continue
			}
			for _, e := range md.Characters.Edges {
				p.media.Characters = append(p.media.Characters, *e.toCharacter())
			}
			next = md.Characters.PageInfo.HasNextPage
		}
		if next && p.next < p.last {
			p.next++
			more = append(more, p)
		}
	}
	return more, nil
}
