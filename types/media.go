package types

import (
	"encoding/json"
	"time"
)

type MetricsHook func(source string, op string, dur time.Duration, err error)

type Date struct {
	Year  *int
	Month *int
	Day   *int
}

type Score struct {
	Mean       float64
	Rank       float64
	Popularity float64
}

type Studio struct {
	ID   string
	Name string
}

type Image struct {
	URL    string
	Width  int
	Height int
}

type Title struct {
	Text     string
	Language string
	Kind     TitleKind
}

type ExternalID struct {
	Source       string
	ID           string
	ResourceType ResourceType
}

type Relation struct {
	TargetID     string
	TargetSource string
	TargetType   string
	Kind         RelationKind
	// TargetMalID is the target's MAL id when the source knows it ("" otherwise).
	// AniList carries idMal on relation nodes; consumers that fetch the target
	// from MAL need it, since TargetID is the AniList id on AniList relations.
	TargetMalID string
	// TargetFormat is the source's verbatim format string ("NOVEL", "TV", ...)
	// when supplied. It separates a light novel from a manga, which TargetType
	// alone ("MANGA" on AniList) cannot.
	TargetFormat string
}

// Media is the canonical, source-independent response shape. It carries every
// field mono-crawler's domain.Media persists, so a Normalize pass can map 1:1.
type Media struct {
	Titles            []Title
	Synopsis          string
	StartDate         *Date
	EndDate           *Date
	Status            MediaStatus
	MediaType         MediaType
	EpisodeCount      int
	ChapterCount      int
	VolumeCount       int
	DurationMin       int
	Score             *Score
	Genres            []string
	Studios           []Studio
	Images            map[ImageSize]Image
	Relations         []Relation
	ExternalIDs       []ExternalID
	Season            Season
	Year              int
	NumListUsers      int
	NumScoringUsers   int
	NumFavorites      int
	Nsfw              string
	Rating            string
	Background        string
	SourceField       string
	MediaTypeField    string
	StatusField       string
	AverageEpDur      int
	Broadcast         json.RawMessage
	Pictures          json.RawMessage
	AlternativeTitles json.RawMessage
	Rankings          json.RawMessage
	Stats             json.RawMessage
	Authors           []Author
	Magazines         []Magazine
	Raw               map[string]any
}

type Author struct {
	ID           string
	FirstName    string
	LastName     string
	Role         string
	ImageMedium  string
	ImageLarge   string
	NumFavorites int
}

type Magazine struct {
	ID   string
	Name string
}
