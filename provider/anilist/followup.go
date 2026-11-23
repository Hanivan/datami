package anilist

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// followUps fetches the remaining pages of every pending connection in
// rounds. Each round asks for the next page of every connection still
// pending, packed followUpAliases per request across works. A connection
// drops out when its page reports no next page, its Media answers null, or it
// reaches its cap. Any failed request, or a reply carrying GraphQL errors,
// fails the call. Requests are counted under OpFollowUps: one mixes cast and
// staff pages of several works.
func (c *Client) followUps(ctx context.Context, pend []pending) error {
	for len(pend) > 0 {
		var keep []pending
		for start := 0; start < len(pend); start += followUpAliases {
			chunk := pend[start:min(start+followUpAliases, len(pend))]
			more, err := c.followUpRequest(ctx, chunk)
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
func (c *Client) followUpRequest(ctx context.Context, chunk []pending) ([]pending, error) {
	const op = OpFollowUps
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
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, helpers.Wrap(types.SourceAniList, op, fmt.Errorf("decode: %w", err))
	}
	// AniList answers 200 with "errors" when it rejects part of the query
	// (complexity, an internal fault); the affected aliases come back null.
	// Treating that as "no more pages" would silently truncate a roster.
	if len(resp.Errors) > 0 {
		return nil, helpers.Wrap(types.SourceAniList, op, fmt.Errorf("graphql: %s", resp.Errors[0].Message))
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
