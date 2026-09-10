package otakotaku

import (
	"strconv"

	"github.com/linkeunid/datami/types"
)

func (r otakotakuResponse) toMedia() *types.Media {
	m := &types.Media{
		Titles: []types.Title{{Text: r.Data.Title}},
		Raw: map[string]any{
			"otakotaku": r.Data.ID,
			"title":     r.Data.Title,
		},
	}
	if r.Data.MalID != nil {
		m.ExternalIDs = append(m.ExternalIDs, types.ExternalID{Source: types.SourceMAL, ID: strconv.Itoa(*r.Data.MalID)})
	}
	if r.Data.AnidbID != nil {
		m.ExternalIDs = append(m.ExternalIDs, types.ExternalID{Source: types.SourceAniDB, ID: strconv.Itoa(*r.Data.AnidbID)})
	}
	if r.Data.AnnID != nil {
		m.ExternalIDs = append(m.ExternalIDs, types.ExternalID{Source: types.SourceAnimeNewsNetwork, ID: strconv.Itoa(*r.Data.AnnID)})
	}
	if r.Data.ApID != nil {
		m.ExternalIDs = append(m.ExternalIDs, types.ExternalID{Source: types.SourceAnimePlanet, ID: strconv.Itoa(*r.Data.ApID)})
	}
	return m
}
