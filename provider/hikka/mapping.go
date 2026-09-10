package hikka

import (
	"strconv"

	"github.com/linkeunid/datami/types"
)

func (i hikkaItem) toMedia() *types.Media {
	m := &types.Media{Raw: map[string]any{
		"slug":     i.Slug,
		"title_ja": i.TitleJA,
		"title_en": i.TitleEN,
		"title_ua": i.TitleUA,
	}}
	if i.TitleEN != "" {
		m.Titles = append(m.Titles, types.Title{Text: i.TitleEN, Language: "en"})
	}
	if i.TitleJA != "" {
		m.Titles = append(m.Titles, types.Title{Text: i.TitleJA, Language: "ja"})
	}
	if i.TitleUA != "" {
		m.Titles = append(m.Titles, types.Title{Text: i.TitleUA, Language: "ua"})
	}
	if m.Titles == nil && i.TitleJA == "" {
		m.Titles = []types.Title{{Text: i.Slug}}
	}
	if i.MalID > 0 {
		m.ExternalIDs = []types.ExternalID{{Source: types.SourceMAL, ID: strconv.Itoa(i.MalID)}}
	}
	return m
}
