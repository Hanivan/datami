package mal

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

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
		m.Relations = append(m.Relations, types.Relation{
			TargetID:     strconv.Itoa(r.Node.ID),
			TargetSource: types.SourceMAL,
			TargetType:   string(types.ResourceAnime),
			Kind:         types.RelationKind(helpers.NormalizeRelation(r.RelationType)),
		})
	}

	for _, r := range a.RelatedManga {
		m.Relations = append(m.Relations, types.Relation{
			TargetID:     strconv.Itoa(r.Node.ID),
			TargetSource: types.SourceMAL,
			TargetType:   string(types.ResourceManga),
			Kind:         types.RelationKind(helpers.NormalizeRelation(r.RelationType)),
		})
	}

	m.StartDate = helpers.ParseDate(a.StartDate)
	m.EndDate = helpers.ParseDate(a.EndDate)

	m.Broadcast = verbatimJSON(a.Broadcast)
	m.Pictures = verbatimPictures(a.Pictures)
	if a.AlternativeTitles != nil {
		m.AlternativeTitles = verbatimAltTitles(*a.AlternativeTitles)
	}

	m.Raw = helpers.RawPayload(a)

	return m
}

// verbatimJSON re-marshals a struct into json.RawMessage so the upsert can
// persist the exact upstream shape. Returns nil when the struct holds no real
// data so the persisted column stays SQL NULL.
func verbatimJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(b, &decoded); err != nil {
		return nil
	}
	if jsonLeafEmpty(decoded) {
		return nil
	}
	return json.RawMessage(b)
}

func jsonLeafEmpty(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case bool:
		return !x
	case float64:
		return x == 0
	case []any:
		for _, e := range x {
			if !jsonLeafEmpty(e) {
				return false
			}
		}
		return true
	case map[string]any:
		for _, e := range x {
			if !jsonLeafEmpty(e) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func verbatimPictures(p []picture) json.RawMessage {
	if len(p) == 0 {
		return nil
	}
	return verbatimJSON(p)
}

func verbatimAltTitles(alt altTitles) json.RawMessage {
	if alt.Synonyms == nil {
		alt.Synonyms = []string{}
	}
	return verbatimJSON(alt)
}

func (m manga) toMedia() *types.Media {
	media := &types.Media{
		ChapterCount:    m.NumChapters,
		VolumeCount:     m.NumVolumes,
		Synopsis:        m.Synopsis,
		NumListUsers:    m.NumListUsers,
		NumScoringUsers: m.NumScoringUsers,
		NumFavorites:    m.NumFavorites,
		Nsfw:            m.NSFW,
		Background:      m.Background,
		MediaTypeField:  m.MediaType,
		StatusField:     m.Status,
		ExternalIDs: []types.ExternalID{
			{Source: types.SourceMAL, ID: strconv.Itoa(m.ID), ResourceType: types.ResourceManga},
		},
	}

	if mapped, ok := malStatusMap[m.Status]; ok {
		media.Status = types.MediaStatus(mapped)
	} else if m.Status != "" {
		media.Status = types.MediaStatus(strings.ReplaceAll(m.Status, "_", " "))
	}

	if mapped, ok := malMediaTypeMap[m.MediaType]; ok {
		media.MediaType = types.MediaType(mapped)
	} else if m.MediaType != "" {
		media.MediaType = types.MediaType(strings.ToLower(m.MediaType))
	}

	if m.Mean > 0 {
		media.Score = &types.Score{Mean: m.Mean, Rank: float64(m.Rank), Popularity: float64(m.Popularity)}
	}

	media.Titles = buildTitles(m.Title, m.AlternativeTitles)

	if m.MainPicture != nil {
		media.Images = map[types.ImageSize]types.Image{
			types.SizeMedium: {URL: m.MainPicture.Medium},
			types.SizeLarge:  {URL: m.MainPicture.Large},
		}
	}

	for _, g := range m.Genres {
		media.Genres = append(media.Genres, g.Name)
	}

	for _, r := range m.RelatedManga {
		media.Relations = append(media.Relations, types.Relation{
			TargetID:     strconv.Itoa(r.Node.ID),
			TargetSource: types.SourceMAL,
			TargetType:   string(types.ResourceManga),
			Kind:         types.RelationKind(helpers.NormalizeRelation(r.RelationType)),
		})
	}

	for _, r := range m.RelatedAnime {
		media.Relations = append(media.Relations, types.Relation{
			TargetID:     strconv.Itoa(r.Node.ID),
			TargetSource: types.SourceMAL,
			TargetType:   string(types.ResourceAnime),
			Kind:         types.RelationKind(helpers.NormalizeRelation(r.RelationType)),
		})
	}

	media.StartDate = helpers.ParseDate(m.StartDate)
	media.EndDate = helpers.ParseDate(m.EndDate)

	for _, au := range m.Authors {
		author := types.Author{
			ID:           strconv.Itoa(au.Node.ID),
			FirstName:    au.Node.FirstName,
			LastName:     au.Node.LastName,
			Role:         au.Role,
			NumFavorites: au.Node.NumFavorites,
		}
		if au.Node.MainPicture != nil {
			author.ImageMedium = au.Node.MainPicture.Medium
			author.ImageLarge = au.Node.MainPicture.Large
		}
		media.Authors = append(media.Authors, author)
	}

	for _, sg := range m.Serialization {
		media.Magazines = append(media.Magazines, types.Magazine{
			ID:   strconv.Itoa(sg.Node.ID),
			Name: sg.Node.Name,
		})
	}

	media.Pictures = verbatimPictures(m.Pictures)
	if m.AlternativeTitles != nil {
		media.AlternativeTitles = verbatimAltTitles(*m.AlternativeTitles)
	}

	media.Raw = helpers.RawPayload(m)

	return media
}

func (a animeCharacter) toCharacter() *types.Character {
	c := &types.Character{
		ID:        strconv.Itoa(a.Node.ID),
		Biography: a.Node.Biography,
		Roles: []types.CharacterRole{
			{MediaID: "", Role: types.CharacterRoleType(strings.ToLower(a.Role))},
		},
	}
	if a.Node.FirstName != "" || a.Node.LastName != "" {
		c.Names = append(c.Names, a.Node.FirstName+" "+a.Node.LastName)
	}
	if a.Node.MainPicture != nil {
		c.Image = a.Node.MainPicture.Medium
	}
	return c
}

func (p person) toPerson() *types.Person {
	out := &types.Person{
		ID:        strconv.Itoa(p.ID),
		Name:      p.FirstName + " " + p.LastName,
		Birthday:  p.Birthday,
		Biography: p.More,
	}
	if p.MainPicture != nil {
		out.Image = p.MainPicture.Medium
	}
	return out
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
