package mal

import (
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

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
		media.Relations = append(media.Relations, relationTarget(r, types.ResourceManga))
	}

	for _, r := range m.RelatedAnime {
		media.Relations = append(media.Relations, relationTarget(r, types.ResourceAnime))
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
	media.Stats = verbatimJSON(m.Statistics)

	media.Raw = helpers.RawPayload(m)

	return media
}
