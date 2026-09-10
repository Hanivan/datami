package anilist

import (
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

func (m media) toMedia() *types.Media {
	result := &types.Media{
		Synopsis:     helpers.DerefString(m.Description),
		EpisodeCount: helpers.DerefInt(m.Episodes),
		ChapterCount: helpers.DerefInt(m.Chapters),
		DurationMin:  helpers.DerefInt(m.Duration),
	}

	if m.Status != nil {
		if mapped, ok := anilistStatusMap[*m.Status]; ok {
			result.Status = types.MediaStatus(mapped)
		} else if *m.Status != "" {
			result.Status = types.MediaStatus(strings.ToLower(*m.Status))
		}
	}

	if m.Format != nil {
		if mapped, ok := anilistFormatMap[*m.Format]; ok {
			result.MediaType = types.MediaType(mapped)
		} else if *m.Format != "" {
			result.MediaType = types.MediaType(strings.ToLower(*m.Format))
		}
	}

	if m.AverageScore != nil || m.MeanScore != nil || m.Popularity != nil {
		result.Score = &types.Score{}
		if m.AverageScore != nil {
			result.Score.Mean = float64(*m.AverageScore)
		}
		if m.MeanScore != nil {
			result.Score.Mean = float64(*m.MeanScore)
		}
		if m.Popularity != nil {
			result.Score.Popularity = float64(*m.Popularity)
		}
	}

	result.Titles = buildTitles(m.Title, m.Synonyms)

	if m.CoverImage != nil {
		result.Images = map[types.ImageSize]types.Image{
			types.SizeMedium: {URL: m.CoverImage.Medium},
			types.SizeLarge:  {URL: m.CoverImage.Large},
		}
	}

	result.Genres = m.Genres

	if m.Studios != nil {
		for _, s := range m.Studios.Nodes {
			result.Studios = append(result.Studios, types.Studio{
				ID:   strconv.Itoa(s.ID),
				Name: s.Name,
			})
		}
	}

	if m.Relations != nil {
		for _, e := range m.Relations.Edges {
			kind := types.ResourceAnime
			if e.Node.Type != nil {
				kind = types.ResourceType(strings.ToLower(*e.Node.Type))
			}
			rel := types.Relation{
				TargetID:     strconv.Itoa(e.Node.ID),
				TargetSource: types.SourceAniList,
				TargetType:   string(kind),
				Kind:         types.RelationKind(helpers.NormalizeRelation(e.RelationType)),
			}
			if e.Node.IDMal != nil {
				rel.TargetMalID = strconv.Itoa(*e.Node.IDMal)
			}
			if e.Node.Format != nil {
				rel.TargetFormat = *e.Node.Format
			}
			result.Relations = append(result.Relations, rel)
		}
	}

	result.StartDate = toDate(m.StartDate)
	result.EndDate = toDate(m.EndDate)

	result.ExternalIDs = []types.ExternalID{
		{Source: types.SourceAniList, ID: strconv.Itoa(m.ID), ResourceType: types.ResourceAnime},
	}
	if m.MALID != nil {
		result.ExternalIDs = append(result.ExternalIDs, types.ExternalID{
			Source:       types.SourceMAL,
			ID:           strconv.Itoa(*m.MALID),
			ResourceType: types.ResourceAnime,
		})
	}

	result.Raw = helpers.RawPayload(m)

	return result
}

func buildTitles(t mediaTitle, synonyms []string) []types.Title {
	titles := []types.Title{}
	if t.Romaji != nil && *t.Romaji != "" {
		titles = append(titles, types.Title{Text: *t.Romaji, Language: "", Kind: types.KindRomaji})
	}
	if t.English != nil && *t.English != "" && *t.English != helpers.DerefString(t.Romaji) {
		titles = append(titles, types.Title{Text: *t.English, Language: "en", Kind: types.KindEnglish})
	}
	if t.Native != nil && *t.Native != "" {
		titles = append(titles, types.Title{Text: *t.Native, Language: "ja", Kind: types.KindNative})
	}
	for _, s := range synonyms {
		if s != "" {
			titles = append(titles, types.Title{Text: s, Language: "", Kind: types.KindSynonym})
		}
	}
	return titles
}

func toDate(f *fuzzyDate) *types.Date {
	if f == nil {
		return nil
	}
	d := &types.Date{}
	if f.Year != nil {
		y := *f.Year
		d.Year = &y
	}
	if f.Month != nil {
		m := *f.Month
		d.Month = &m
	}
	if f.Day != nil {
		day := *f.Day
		d.Day = &day
	}
	return d
}

func (e characterEdge) toCharacter() *types.Character {
	c := &types.Character{
		ID:        strconv.Itoa(e.Node.ID),
		Biography: helpers.DerefString(e.Node.Description),
		Roles: []types.CharacterRole{
			{MediaID: "", Role: types.CharacterRoleType(strings.ToLower(e.Role))},
		},
	}
	name := e.Node.Name.Full
	if name == "" {
		name = strings.TrimSpace(e.Node.Name.First + " " + e.Node.Name.Last)
	}
	c.Names = append(c.Names, name)
	if e.Node.Image.Large != "" || e.Node.Image.Medium != "" {
		c.Image = e.Node.Image.Large
		if c.Image == "" {
			c.Image = e.Node.Image.Medium
		}
	}
	return c
}

func (e staffEdge) toPerson() *types.Person {
	name := e.Node.Name.Full
	if name == "" {
		name = e.Node.Name.Native
	}
	img := e.Node.Image.Large
	if img == "" {
		img = e.Node.Image.Medium
	}
	return &types.Person{
		ID:    strconv.Itoa(e.Node.ID),
		Name:  name,
		Image: img,
	}
}
