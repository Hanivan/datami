package kitsu

type anime struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Attributes animeAttrs `json:"attributes"`
}

type animeAttrs struct {
	Slug           string       `json:"slug"`
	Synopsis       string       `json:"synopsis"`
	StartDate      *string      `json:"startDate"`
	EndDate        *string      `json:"endDate"`
	Status         *string      `json:"status"`
	ShowType       *string      `json:"showType"`
	MangaType      *string      `json:"mangaType"`
	EpisodeCount   *int         `json:"episodeCount"`
	EpisodeLength  *int         `json:"episodeLength"`
	ChapterCount   *int         `json:"chapterCount"`
	VolumeCount    *int         `json:"volumeCount"`
	AverageRating  *float64     `json:"averageRating"`
	UserCount      *int         `json:"userCount"`
	FavoritesCount *int         `json:"favoritesCount"`
	Genres         []string     `json:"genres"`
	Titles         *kitsuTitles `json:"titles"`
	PosterImage    *posterImage `json:"posterImage"`
}

type kitsuTitles struct {
	EN   *string `json:"en"`
	ENJp *string `json:"en_jp"`
	JAJp *string `json:"ja_jp"`
}

type posterImage struct {
	Medium   string `json:"medium"`
	Original string `json:"original"`
}

type searchResponse struct {
	Data []anime `json:"data"`
}
