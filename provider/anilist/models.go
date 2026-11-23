package anilist

import (
	"encoding/json"

	"github.com/Hanivan/datami/types"
)

type media struct {
	ID              int             `json:"id"`
	Type            *string         `json:"type,omitempty"`
	Title           mediaTitle      `json:"title"`
	Description     *string         `json:"description,omitempty"`
	StartDate       *fuzzyDate      `json:"startDate,omitempty"`
	EndDate         *fuzzyDate      `json:"endDate,omitempty"`
	Status          *string         `json:"status,omitempty"`
	Format          *string         `json:"format,omitempty"`
	Episodes        *int            `json:"episodes,omitempty"`
	Duration        *int            `json:"duration,omitempty"`
	Chapters        *int            `json:"chapters,omitempty"`
	Volumes         *int            `json:"volumes,omitempty"`
	AverageScore    *int            `json:"averageScore,omitempty"`
	MeanScore       *int            `json:"meanScore,omitempty"`
	Popularity      *int            `json:"popularity,omitempty"`
	Favourites      *int            `json:"favourites,omitempty"`
	Genres          []string        `json:"genres,omitempty"`
	Tags            []tag           `json:"tags,omitempty"`
	Studios         *studioConn     `json:"studios,omitempty"`
	CoverImage      *coverImage     `json:"coverImage,omitempty"`
	BannerImage     *string         `json:"bannerImage,omitempty"`
	Relations       *relationConn   `json:"relations,omitempty"`
	Synonyms        []string        `json:"synonyms,omitempty"`
	MALID           *int            `json:"idMal,omitempty"`
	Source          *string         `json:"source,omitempty"`
	CountryOfOrigin *string         `json:"countryOfOrigin,omitempty"`
	IsAdult         *bool           `json:"isAdult,omitempty"`
	ExternalLinks   []externalLink  `json:"externalLinks,omitempty"`
	Rankings        json.RawMessage `json:"rankings,omitempty"`
	Stats           json.RawMessage `json:"stats,omitempty"`
}

type tag struct {
	Name             string `json:"name"`
	Category         string `json:"category"`
	Rank             *int   `json:"rank,omitempty"`
	IsGeneralSpoiler bool   `json:"isGeneralSpoiler"`
	IsMediaSpoiler   bool   `json:"isMediaSpoiler"`
	IsAdult          bool   `json:"isAdult"`
}

type externalLink struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	Site       string `json:"site"`
	SiteID     *int   `json:"siteId,omitempty"`
	Type       string `json:"type"`
	Language   string `json:"language"`
	Notes      string `json:"notes"`
	IsDisabled bool   `json:"isDisabled"`
}

type mediaTitle struct {
	Romaji  *string `json:"romaji,omitempty"`
	English *string `json:"english,omitempty"`
	Native  *string `json:"native,omitempty"`
}

type fuzzyDate struct {
	Year  *int `json:"year,omitempty"`
	Month *int `json:"month,omitempty"`
	Day   *int `json:"day,omitempty"`
}

type studioConn struct {
	Edges []studioEdge `json:"edges"`
}

type studioEdge struct {
	IsMain bool       `json:"isMain"`
	Node   studioNode `json:"node"`
}

type studioNode struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	IsAnimationStudio bool   `json:"isAnimationStudio"`
}

type coverImage struct {
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type relationConn struct {
	Edges []relationEdge `json:"edges"`
}

type relationEdge struct {
	RelationType string `json:"relationType"`
	Node         struct {
		ID     int        `json:"id"`
		Type   *string    `json:"type,omitempty"`
		IDMal  *int       `json:"idMal,omitempty"`
		Format *string    `json:"format,omitempty"`
		Title  mediaTitle `json:"title"`
	} `json:"node"`
}

// pending is one cast or staff connection with pages left to fetch: next is
// the page to ask for, last the connection's page cap.
type pending struct {
	media   *types.Media
	mediaID int
	staff   bool
	next    int
	last    int
}

// Aliases keep the anilist.* names; the types live in types/tag.go.
type (
	TagsLinks    = types.TagsLinks
	Tag          = types.Tag
	ExternalLink = types.ExternalLink
)

type mediaResponse struct {
	Data struct {
		Media *media `json:"Media"`
	} `json:"Data"`
}

type searchResponse struct {
	Data struct {
		Page struct {
			Media []media `json:"Media"`
		} `json:"Page"`
	} `json:"Data"`
}
