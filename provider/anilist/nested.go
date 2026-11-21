package anilist

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Hanivan/datami/types"
)

const (
	// maxPages is the default total page cap per connection (embedded +
	// follow-up), and the fixed cap of the standalone GetCharacters/GetStaff walk.
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

// pageCap is a connection's total page cap: max, or maxPages when unset.
func pageCap(max int) int {
	if max > 0 {
		return max
	}
	return maxPages
}

// embedPages appends c1..cN / s1..sN aliased connections to a Media query,
// which ends with the Media close and the query close.
func (c *Client) embedPages(query string, o types.MediaOpts) string {
	nc, ns := min(o.CharacterPages, pageCap(o.MaxCharacterPages)), min(o.StaffPages, pageCap(o.MaxStaffPages))
	if nc <= 0 && ns <= 0 {
		return query
	}
	var b strings.Builder
	b.WriteString(strings.TrimSuffix(strings.TrimSuffix(query, " }"), " }"))
	for i := 1; i <= nc; i++ {
		fmt.Fprintf(&b, " c%d: characters(page: %d, perPage: %d, %s) %s", i, i, pageSize, charSort, c.charBody())
	}
	for i := 1; i <= ns; i++ {
		fmt.Fprintf(&b, " s%d: staff(page: %d, perPage: %d, %s) %s", i, i, pageSize, staffSort, staffBody)
	}
	b.WriteString(" } }")
	return b.String()
}

// mergeEmbedded decodes the aliased pages of a single-media response into m
// and returns the connections that still have pages to fetch.
func mergeEmbedded(body []byte, mediaID int, o types.MediaOpts, m *types.Media) ([]pending, error) {
	var raw struct {
		Data struct {
			Media map[string]json.RawMessage `json:"Media"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, &types.Error{Source: types.SourceAniList, Op: OpGetCharacters, Err: fmt.Errorf("decode: %w", err)}
	}
	return mergeEmbeddedMap(raw.Data.Media, mediaID, o, m)
}

// mergeEmbeddedMap is mergeEmbedded over one already-split media object.
func mergeEmbeddedMap(media map[string]json.RawMessage, mediaID int, o types.MediaOpts, m *types.Media) ([]pending, error) {
	var pend []pending
	capC, capS := pageCap(o.MaxCharacterPages), pageCap(o.MaxStaffPages)
	nc, ns := min(o.CharacterPages, capC), min(o.StaffPages, capS)

	for i := 1; i <= nc; i++ {
		var conn charConn
		if r, ok := media["c"+strconv.Itoa(i)]; ok && string(r) != "null" {
			if err := json.Unmarshal(r, &conn); err != nil {
				return nil, &types.Error{Source: types.SourceAniList, Op: OpGetCharacters, Err: fmt.Errorf("decode: %w", err)}
			}
		}
		for _, e := range conn.Edges {
			m.Characters = append(m.Characters, *e.toCharacter())
		}
		if !conn.PageInfo.HasNextPage {
			break
		}
		if i == nc && nc < capC {
			pend = append(pend, pending{media: m, mediaID: mediaID, next: nc + 1, last: capC})
		}
	}

	for i := 1; i <= ns; i++ {
		var conn staffConn
		if r, ok := media["s"+strconv.Itoa(i)]; ok && string(r) != "null" {
			if err := json.Unmarshal(r, &conn); err != nil {
				return nil, &types.Error{Source: types.SourceAniList, Op: OpGetStaff, Err: fmt.Errorf("decode: %w", err)}
			}
		}
		for _, e := range conn.Edges {
			m.Staff = append(m.Staff, *e.toPerson())
		}
		if !conn.PageInfo.HasNextPage {
			break
		}
		if i == ns && ns < capS {
			pend = append(pend, pending{media: m, mediaID: mediaID, staff: true, next: ns + 1, last: capS})
		}
	}
	return pend, nil
}
