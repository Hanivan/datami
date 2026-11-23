package anidb

import (
	"strconv"
	"strings"
	"time"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// toMedia maps a decoded anime; now drives status derivation.
func (a anime) toMedia(now time.Time) *types.Media {
	m := &types.Media{
		Synopsis:     a.Description,
		EpisodeCount: a.EpisodeCount,
		Titles:       buildTitles(a.Titles),
		StartDate:    helpers.ParseDate(a.StartDate),
		EndDate:      helpers.ParseDate(a.EndDate),
		ExternalIDs: []types.ExternalID{
			{Source: types.SourceAniDB, ID: strconv.Itoa(a.ID), ResourceType: types.ResourceAnime},
		},
		Raw: helpers.RawPayload(a),
	}
	m.Status = deriveStatus(m.StartDate, m.EndDate, now)

	if mapped, ok := anidbTypeMap[a.Type]; ok {
		m.MediaType = types.MediaType(mapped)
	} else if a.Type != "" {
		m.MediaType = types.MediaType(helpers.NormalizeRelation(a.Type))
	}

	if a.Ratings.Permanent.Value > 0 {
		m.Score = &types.Score{Mean: a.Ratings.Permanent.Value}
	}

	if a.Picture != "" {
		img := types.Image{URL: imageBaseURL + a.Picture}
		m.Images = map[types.ImageSize]types.Image{types.SizeOriginal: img, types.SizeLarge: img}
	}

	for _, t := range a.Tags {
		if t.Weight >= genreMinWeight {
			m.Genres = append(m.Genres, t.Name)
		}
	}

	for _, c := range a.Creators {
		if c.Type == "Animation Work" {
			m.Studios = append(m.Studios, types.Studio{ID: strconv.Itoa(c.ID), Name: strings.TrimSpace(c.Name)})
		}
	}

	for _, r := range a.Related {
		kind, ok := relationMap[r.Type]
		if !ok {
			kind = types.RelationOther
		}
		m.Relations = append(m.Relations, types.Relation{
			TargetID:     strconv.Itoa(r.ID),
			TargetSource: types.SourceAniDB,
			TargetType:   string(types.ResourceAnime),
			Kind:         kind,
			TargetTitle:  strings.TrimSpace(r.Name),
		})
	}

	// Unknown resource types stay in Raw only.
	for _, r := range a.Resources {
		src, ok := resourceSourceMap[r.Type]
		if !ok {
			continue
		}
		for _, e := range r.Entities {
			if len(e.Identifiers) > 0 && e.Identifiers[0] != "" {
				m.ExternalIDs = append(m.ExternalIDs, types.ExternalID{Source: src, ID: e.Identifiers[0], ResourceType: types.ResourceAnime})
			}
		}
	}

	return m
}

func buildTitles(t titles) []types.Title {
	out := []types.Title{}
	for _, e := range t.Title {
		var kind types.TitleKind
		switch e.Type {
		case "main":
			kind = types.KindCanonical
		case "official":
			k, ok := titleKindMap[e.Lang]
			if !ok {
				k = types.KindSynonym
			}
			kind = k
		case "short":
			kind = types.KindShort
		default:
			kind = types.KindSynonym
		}
		out = append(out, types.Title{Text: strings.TrimSpace(e.Text), Language: e.Lang, Kind: kind})
	}
	// Main title first.
	for i, x := range out {
		if x.Kind == types.KindCanonical {
			copy(out[1:i+1], out[:i])
			out[0] = x
			break
		}
	}
	return out
}

// deriveStatus: end in past -> finished; start in future -> planned;
// start past -> ongoing; no start -> "".
func deriveStatus(start, end *types.Date, now time.Time) types.MediaStatus {
	if t, ok := dateStart(end); ok && t.Before(now) {
		return types.StatusFinished
	}
	t, ok := dateStart(start)
	switch {
	case !ok:
		return ""
	case t.After(now):
		return types.StatusPlanned
	default:
		return types.StatusOngoing
	}
}

// dateStart treats missing month/day as the start of the period.
func dateStart(d *types.Date) (time.Time, bool) {
	if d == nil || d.Year == nil {
		return time.Time{}, false
	}
	month, day := 1, 1
	if d.Month != nil {
		month = *d.Month
	}
	if d.Day != nil {
		day = *d.Day
	}
	return time.Date(*d.Year, time.Month(month), day, 0, 0, 0, 0, time.UTC), true
}

func cdnImage(picture string) string {
	if picture == "" {
		return ""
	}
	return imageBaseURL + picture
}
