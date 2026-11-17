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
	ID         int
	URL        string
	Site       string
	SiteID     *int
	Type       string
	Language   string
	Notes      string
	IsDisabled bool
}
