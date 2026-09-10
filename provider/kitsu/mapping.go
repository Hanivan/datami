package kitsu

import (
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

func (a anime) toMedia() *types.Media {
	attrs := a.Attributes
	m := &types.Media{
		Synopsis:     attrs.Synopsis,
		EpisodeCount: helpers.DerefInt(attrs.EpisodeCount),
		ChapterCount: helpers.DerefInt(attrs.ChapterCount),
		DurationMin:  helpers.DerefInt(attrs.EpisodeLength),
		Genres:       attrs.Genres,
		ExternalIDs: []types.ExternalID{
			{Source: types.SourceKitsu, ID: a.ID, ResourceType: types.ResourceAnime},
		},
		Raw: make(map[string]any),
	}

	if attrs.Status != nil {
		if mapped, ok := kitsuStatusMap[*attrs.Status]; ok {
			m.Status = types.MediaStatus(mapped)
		} else if *attrs.Status != "" {
			m.Status = types.MediaStatus(strings.ToLower(*attrs.Status))
		}
	}

	if attrs.ShowType != nil {
		if mapped, ok := kitsuShowTypeMap[*attrs.ShowType]; ok {
			m.MediaType = types.MediaType(mapped)
		} else if *attrs.ShowType != "" {
			m.MediaType = types.MediaType(strings.ToLower(*attrs.ShowType))
		}
	}

	if attrs.AverageRating != nil || attrs.UserCount != nil {
		m.Score = &types.Score{}
		if attrs.AverageRating != nil {
			m.Score.Mean = *attrs.AverageRating
		}
		if attrs.UserCount != nil {
			m.Score.Popularity = float64(*attrs.UserCount)
		}
	}

	m.Titles = buildTitles(attrs.Titles)

	if attrs.PosterImage != nil {
		m.Images = map[types.ImageSize]types.Image{
			types.SizeMedium: {URL: attrs.PosterImage.Medium},
			types.SizeLarge:  {URL: attrs.PosterImage.Original},
		}
	}

	m.StartDate = helpers.ParseDate(helpers.DerefString(attrs.StartDate))
	m.EndDate = helpers.ParseDate(helpers.DerefString(attrs.EndDate))

	return m
}

func buildTitles(t *kitsuTitles) []types.Title {
	titles := []types.Title{}
	if t == nil {
		return titles
	}
	if t.ENJp != nil && *t.ENJp != "" {
		titles = append(titles, types.Title{Text: *t.ENJp, Language: "", Kind: types.KindRomaji})
	}
	if t.EN != nil && *t.EN != "" && *t.EN != helpers.DerefString(t.ENJp) {
		titles = append(titles, types.Title{Text: *t.EN, Language: "en", Kind: types.KindEnglish})
	}
	if t.JAJp != nil && *t.JAJp != "" {
		titles = append(titles, types.Title{Text: *t.JAJp, Language: "ja", Kind: types.KindNative})
	}
	return titles
}
