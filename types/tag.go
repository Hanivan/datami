package types

// TagsLinks holds a media's tags + external links from one fetch.
type TagsLinks struct {
	Tags          []Tag
	ExternalLinks []ExternalLink
}

type Tag struct {
	Name             string
	Category         string
	Rank             *int
	IsGeneralSpoiler bool
	IsMediaSpoiler   bool
	IsAdult          bool
}

type ExternalLink struct {
	ID         int    `json:"id"`
	URL        string `json:"url"`
	Site       string `json:"site"`
	SiteID     *int   `json:"site_id,omitempty"`
	Type       string `json:"type"`
	Language   string `json:"language"`
	Notes      string `json:"notes"`
	IsDisabled bool   `json:"is_disabled"`
}
