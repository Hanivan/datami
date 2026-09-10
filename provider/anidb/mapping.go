package anidb

import (
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

func (a anime) toMedia() *types.Media {
	m := &types.Media{
		Synopsis:     a.Description,
		EpisodeCount: a.Episodes,
		ExternalIDs: []types.ExternalID{
			{Source: types.SourceAniDB, ID: strconv.Itoa(a.ID), ResourceType: types.ResourceAnime},
		},
		Raw: make(map[string]any),
	}

	if mapped, ok := anidbTypeMap[a.Type]; ok {
		m.MediaType = types.MediaType(mapped)
	} else if a.Type != "" {
		m.MediaType = types.MediaType(strings.ToLower(strings.ReplaceAll(a.Type, " ", "_")))
	}

	m.Titles = buildTitles(a.Titles)

	if a.Picture != "" {
		m.Images = map[types.ImageSize]types.Image{
			types.SizeMedium: {URL: "https://cdn.anidb.net/images/main/" + a.Picture},
			types.SizeLarge:  {URL: "https://cdn.anidb.net/images/main/" + a.Picture},
		}
	}

	for _, t := range a.Tags {
		m.Genres = append(m.Genres, t.Name)
	}

	for _, r := range a.Related {
		m.Relations = append(m.Relations, types.Relation{
			TargetID:     strconv.Itoa(r.ID),
			TargetSource: types.SourceAniDB,
			TargetType:   string(types.ResourceAnime),
			Kind:         types.RelationOther,
		})
	}

	m.StartDate = helpers.ParseDate(a.StartDate)
	m.EndDate = helpers.ParseDate(a.EndDate)

	return m
}

func buildTitles(t titles) []types.Title {
	titles_ := []types.Title{}
	for _, entry := range t.Title {
		kind := types.KindSynonym
		switch entry.Type {
		case "official":
			if entry.Lang == "en" {
				kind = types.KindEnglish
			} else if entry.Lang == "ja" {
				kind = types.KindRomaji
			} else {
				kind = types.KindCanonical
			}
		case "short":
			kind = types.KindShort
		}
		titles_ = append(titles_, types.Title{
			Text:     entry.Text,
			Language: entry.Lang,
			Kind:     kind,
		})
	}
	return titles_
}
