package anilist

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Hanivan/datami/types"
)

const (
	// maxPages caps total pages per connection (embedded + follow-up).
	maxPages = 10
	// pageSize is AniList's per-page ceiling; never send more.
	pageSize = 25

	charSort  = "sort: ROLE"
	staffSort = "sort: RELEVANCE"

	// staffBody is the selection inside a staff connection, shared by the
	// standalone and embedded queries.
	staffBody = `{ pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } primaryOccupations } } }`
)

// charBody is the selection inside a characters connection, shared by the
// standalone and embedded queries.
func (c *Client) charBody() string {
	return `{ pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } ` + c.desc() + ` } voiceActors { id name { first middle last full native alternative } image { medium large } languageV2 } } }`
}

// resolvePages applies PageOpts defaults and the global page cap.
func resolvePages(opts []types.PageOpts) (start, max int) {
	start, max = 1, maxPages
	if len(opts) > 0 {
		if opts[0].StartPage > 0 {
			start = opts[0].StartPage
		}
		if opts[0].MaxPages > 0 {
			max = opts[0].MaxPages
		}
	}
	// The last page fetched never exceeds maxPages, whatever the opts say.
	if room := maxPages - start + 1; max > room {
		max = room
	}
	return start, max
}

// embedPages appends c1..cN / s1..sN aliased connections to a Media query,
// which ends with the Media close and the query close.
func (c *Client) embedPages(query string, o types.MediaOpts) string {
	if o.CharacterPages <= 0 && o.StaffPages <= 0 {
		return query
	}
	var b strings.Builder
	b.WriteString(strings.TrimSuffix(strings.TrimSuffix(query, " }"), " }"))
	for i := 1; i <= min(o.CharacterPages, maxPages); i++ {
		fmt.Fprintf(&b, " c%d: characters(page: %d, perPage: %d, %s) %s", i, i, pageSize, charSort, c.charBody())
	}
	for i := 1; i <= min(o.StaffPages, maxPages); i++ {
		fmt.Fprintf(&b, " s%d: staff(page: %d, perPage: %d, %s) %s", i, i, pageSize, staffSort, staffBody)
	}
	b.WriteString(" } }")
	return b.String()
}

// mergeEmbedded decodes the aliased pages from body into m, then walks the
// remaining pages with a standalone follow-up when the last embedded page
// reported more.
func (c *Client) mergeEmbedded(ctx context.Context, body []byte, mediaID int, o types.MediaOpts, m *types.Media) error {
	var raw struct {
		Data struct {
			Media map[string]json.RawMessage `json:"Media"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return &types.Error{Source: types.SourceAniList, Op: OpGetCharacters, Err: fmt.Errorf("decode: %w", err)}
	}
	return c.mergeEmbeddedMap(ctx, raw.Data.Media, mediaID, o, m)
}

// mergeEmbeddedMap is mergeEmbedded over one already-split media object.
func (c *Client) mergeEmbeddedMap(ctx context.Context, media map[string]json.RawMessage, mediaID int, o types.MediaOpts, m *types.Media) error {
	nc, ns := min(o.CharacterPages, maxPages), min(o.StaffPages, maxPages)

	var chars []characterEdge
	var moreChars bool
	for i := 1; i <= nc; i++ {
		var conn charConn
		if r, ok := media["c"+strconv.Itoa(i)]; ok && string(r) != "null" {
			if err := json.Unmarshal(r, &conn); err != nil {
				return &types.Error{Source: types.SourceAniList, Op: OpGetCharacters, Err: fmt.Errorf("decode: %w", err)}
			}
		}
		chars = append(chars, conn.Edges...)
		moreChars = conn.PageInfo.HasNextPage
		if !moreChars {
			break
		}
	}
	for _, e := range chars {
		m.Characters = append(m.Characters, *e.toCharacter())
	}
	if moreChars && nc < maxPages {
		rest, err := c.charactersBy(ctx, "id", strconv.Itoa(mediaID), "", types.PageOpts{StartPage: nc + 1, MaxPages: maxPages - nc})
		if err != nil {
			return err
		}
		m.Characters = append(m.Characters, rest...)
	}

	var staff []staffEdge
	var moreStaff bool
	for i := 1; i <= ns; i++ {
		var conn staffConn
		if r, ok := media["s"+strconv.Itoa(i)]; ok && string(r) != "null" {
			if err := json.Unmarshal(r, &conn); err != nil {
				return &types.Error{Source: types.SourceAniList, Op: OpGetStaff, Err: fmt.Errorf("decode: %w", err)}
			}
		}
		staff = append(staff, conn.Edges...)
		moreStaff = conn.PageInfo.HasNextPage
		if !moreStaff {
			break
		}
	}
	for _, e := range staff {
		m.Staff = append(m.Staff, *e.toPerson())
	}
	if moreStaff && ns < maxPages {
		rest, err := c.staffBy(ctx, "id", strconv.Itoa(mediaID), "", types.PageOpts{StartPage: ns + 1, MaxPages: maxPages - ns})
		if err != nil {
			return err
		}
		m.Staff = append(m.Staff, rest...)
	}
	return nil
}
