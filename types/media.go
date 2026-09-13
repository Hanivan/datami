package types

import (
	"encoding/json"
	"time"
)

type MetricsHook func(source string, op string, dur time.Duration, err error)

type Date struct {
	Year  *int `json:"year,omitempty"`
	Month *int `json:"month,omitempty"`
	Day   *int `json:"day,omitempty"`
}

type Score struct {
	Mean       float64 `json:"mean"`
	Rank       float64 `json:"rank"`
	Popularity float64 `json:"popularity"`
}

type Studio struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type Title struct {
	Text     string `json:"text"`
	Language string `json:"language,omitempty"`
	Kind     TitleKind
}

type ExternalID struct {
	Source       string `json:"source"`
	ID           string `json:"id"`
	ResourceType ResourceType `json:"resource_type,omitempty"`
}

type Relation struct {
	TargetID     string `json:"target_id"`
	TargetSource string `json:"target_source,omitempty"`
	TargetType   string `json:"target_type,omitempty"`
	Kind         RelationKind
	// TargetMalID is the target's MAL id when the source knows it ("" otherwise).
	// AniList carries idMal on relation nodes; consumers that fetch the target
	// from MAL need it, since TargetID is the AniList id on AniList relations.
	TargetMalID string `json:"target_mal_id,omitempty"`
	// TargetFormat is the source's verbatim format string ("NOVEL", "TV", ...)
	// when supplied. It separates a light novel from a manga, which TargetType
	// alone ("MANGA" on AniList) cannot.
	TargetFormat string `json:"target_format,omitempty"`
	// TargetTitle is the target's display title. Display only — never an
	// identity key, since the same work carries different titles per source.
	TargetTitle string `json:"target_title,omitempty"`
}

// Media is the canonical, source-independent response shape. It carries every
// field mono-crawler's domain.Media persists, so a Normalize pass can map 1:1.
type Media struct {
	Titles            []Title      `json:"titles,omitempty"`
	Synopsis          string       `json:"synopsis,omitempty"`
	StartDate         *Date        `json:"start_date,omitempty"`
	EndDate           *Date        `json:"end_date,omitempty"`
	Status            MediaStatus  `json:"-"`
	MediaType         MediaType    `json:"-"`
	EpisodeCount      int          `json:"episode_count,omitempty"`
	ChapterCount      int          `json:"chapter_count,omitempty"`
	VolumeCount       int          `json:"volume_count,omitempty"`
	DurationMin       int          `json:"duration_min,omitempty"`
	Score             *Score       `json:"score,omitempty"`
	Genres            []string     `json:"genres,omitempty"`
	Studios           []Studio     `json:"studios,omitempty"`
	Images            map[ImageSize]Image `json:"images,omitempty"`
	Relations         []Relation   `json:"relations,omitempty"`
	ExternalIDs       []ExternalID `json:"external_ids,omitempty"`
	Season            Season       `json:"season,omitempty"`
	Year              int          `json:"year,omitempty"`
	NumListUsers      int          `json:"num_list_users,omitempty"`
	NumScoringUsers   int          `json:"num_scoring_users,omitempty"`
	NumFavorites      int          `json:"num_favorites,omitempty"`
	Nsfw              string       `json:"nsfw,omitempty"`
	Rating            string       `json:"rating,omitempty"`
	Background        string       `json:"background,omitempty"`
	SourceField       string       `json:"source_field,omitempty"`
	MediaTypeField    string       `json:"media_type_field,omitempty"`
	StatusField       string       `json:"status_field,omitempty"`
	AverageEpDur      int          `json:"average_ep_dur,omitempty"`
	Broadcast         json.RawMessage `json:"broadcast,omitempty"`
	Pictures          json.RawMessage `json:"pictures,omitempty"`
	AlternativeTitles json.RawMessage `json:"alternative_titles,omitempty"`
	Rankings          json.RawMessage `json:"rankings,omitempty"`
	Stats             json.RawMessage `json:"stats,omitempty"`
	Authors           []Author     `json:"authors,omitempty"`
	Magazines         []Magazine   `json:"magazines,omitempty"`
	Raw               map[string]any `json:"-"`
}

type Author struct {
	ID           string `json:"id"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Role         string `json:"role,omitempty"`
	ImageMedium  string `json:"image_medium,omitempty"`
	ImageLarge   string `json:"image_large,omitempty"`
	NumFavorites int    `json:"num_favorites,omitempty"`
}

type Magazine struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}
