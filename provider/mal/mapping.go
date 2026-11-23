package mal

import (
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// relationTarget builds one relation edge. TargetType is the node's VERBATIM
// MAL media_type ("tv"/"ova"/"light_novel"/...) when the request asked for it
// (related_anime{media_type}); it falls back to fallback ("anime"/"manga", the
// list the entry came from) when MAL omitted it. Consumers classify the target
// medium from this string, so the specific media_type matters: a light_novel
// listed under related_manga is not a manga.
func relationTarget(r relatedEntry, fallback types.ResourceType) types.Relation {
	targetType := r.Node.MediaType
	if targetType == "" {
		targetType = string(fallback)
	}
	return types.Relation{
		TargetID:     strconv.Itoa(r.Node.ID),
		TargetSource: types.SourceMAL,
		TargetType:   targetType,
		TargetTitle:  r.Node.Title,
		Kind:         types.RelationKind(helpers.NormalizeRelation(r.RelationType)),
	}
}

func (a anime) toMedia() *types.Media {
	m := &types.Media{
		Synopsis:        a.Synopsis,
		EpisodeCount:    a.NumEpisodes,
		DurationMin:     a.AverageEpDuration / 60,
		NumListUsers:    a.NumListUsers,
		NumScoringUsers: a.NumScoringUsers,
		NumFavorites:    a.NumFavorites,
		Nsfw:            a.NSFW,
		Rating:          a.Rating,
		Background:      a.Background,
		SourceField:     a.Source,
		MediaTypeField:  a.MediaType,
		StatusField:     a.Status,
		AverageEpDur:    a.AverageEpDuration,
		ExternalIDs: []types.ExternalID{
			{Source: types.SourceMAL, ID: strconv.Itoa(a.ID), ResourceType: types.ResourceAnime},
		},
	}

	if a.StartSeason.Year > 0 {
		m.Year = a.StartSeason.Year
		m.Season = types.Season(strings.ToLower(a.StartSeason.Season))
	}

	if mapped, ok := malStatusMap[a.Status]; ok {
		m.Status = types.MediaStatus(mapped)
	} else if a.Status != "" {
		m.Status = types.MediaStatus(strings.ReplaceAll(a.Status, "_", " "))
	}

	if mapped, ok := malMediaTypeMap[a.MediaType]; ok {
		m.MediaType = types.MediaType(mapped)
	} else if a.MediaType != "" {
		m.MediaType = types.MediaType(strings.ToLower(a.MediaType))
	}

	if a.Mean > 0 {
		m.Score = &types.Score{Mean: a.Mean, Rank: float64(a.Rank), Popularity: float64(a.Popularity)}
	}

	m.Titles = buildTitles(a.Title, a.AlternativeTitles)

	if a.MainPicture != nil {
		m.Images = map[types.ImageSize]types.Image{
			types.SizeMedium: {URL: a.MainPicture.Medium},
			types.SizeLarge:  {URL: a.MainPicture.Large},
		}
	}

	for _, g := range a.Genres {
		m.Genres = append(m.Genres, g.Name)
	}

	for _, s := range a.Studios {
		m.Studios = append(m.Studios, types.Studio{ID: strconv.Itoa(s.ID), Name: s.Name})
	}

	for _, r := range a.RelatedAnime {
		m.Relations = append(m.Relations, relationTarget(r, types.ResourceAnime))
	}

	for _, r := range a.RelatedManga {
		m.Relations = append(m.Relations, relationTarget(r, types.ResourceManga))
	}

	m.StartDate = helpers.ParseDate(a.StartDate)
	m.EndDate = helpers.ParseDate(a.EndDate)

	m.Broadcast = verbatimJSON(a.Broadcast)
	m.Pictures = verbatimPictures(a.Pictures)
	if a.AlternativeTitles != nil {
		m.AlternativeTitles = verbatimAltTitles(*a.AlternativeTitles)
	}
	m.Stats = verbatimJSON(a.Statistics)

	m.Raw = helpers.RawPayload(a)

	return m
}

func buildTitles(main string, alt *altTitles) []types.Title {
	titles := []types.Title{{Text: main, Language: "", Kind: types.KindCanonical}}
	if alt == nil {
		return titles
	}
	if alt.EN != "" && alt.EN != main {
		titles = append(titles, types.Title{Text: alt.EN, Language: "en", Kind: types.KindEnglish})
	}
	if alt.JA != "" {
		titles = append(titles, types.Title{Text: alt.JA, Language: "ja", Kind: types.KindRomaji})
	}
	for _, s := range alt.Synonyms {
		if s != "" && s != main {
			titles = append(titles, types.Title{Text: s, Language: "", Kind: types.KindSynonym})
		}
	}
	return titles
}
