package mal

type manga struct {
	ID                int              `json:"id"`
	Title             string           `json:"title"`
	MainPicture       *picture         `json:"main_picture,omitempty"`
	Synopsis          string           `json:"synopsis,omitempty"`
	StartDate         string           `json:"start_date,omitempty"`
	EndDate           string           `json:"end_date,omitempty"`
	Status            string           `json:"status,omitempty"`
	MediaType         string           `json:"media_type,omitempty"`
	NumChapters       int              `json:"num_chapters,omitempty"`
	NumVolumes        int              `json:"num_volumes,omitempty"`
	Mean              float64          `json:"mean,omitempty"`
	Rank              int              `json:"rank,omitempty"`
	Popularity        int              `json:"popularity,omitempty"`
	NumListUsers      int              `json:"num_list_users,omitempty"`
	NumScoringUsers   int              `json:"num_scoring_users,omitempty"`
	NumFavorites      int              `json:"num_favorites,omitempty"`
	NSFW              string           `json:"nsfw,omitempty"`
	Genres            []genre          `json:"genres,omitempty"`
	AlternativeTitles *altTitles       `json:"alternative_titles,omitempty"`
	RelatedManga      []relatedEntry   `json:"related_manga,omitempty"`
	RelatedAnime      []relatedEntry   `json:"related_anime,omitempty"`
	Authors           []mangaAuthor    `json:"authors,omitempty"`
	Serialization     []serialization  `json:"serialization,omitempty"`
	Pictures          []picture        `json:"pictures,omitempty"`
	Background        string           `json:"background,omitempty"`
	Recommendations   []recommendation `json:"recommendations,omitempty"`
	Statistics        *statistics      `json:"statistics,omitempty"`
	MyListStatus      *listStatus      `json:"my_list_status,omitempty"`
	CreatedAt         string           `json:"created_at,omitempty"`
	UpdatedAt         string           `json:"updated_at,omitempty"`
}

type mangaAuthor struct {
	Node struct {
		ID           int      `json:"id"`
		FirstName    string   `json:"first_name"`
		LastName     string   `json:"last_name"`
		MainPicture  *picture `json:"main_picture,omitempty"`
		NumFavorites int      `json:"num_favorites,omitempty"`
	} `json:"node"`
	Role string `json:"role"`
}

type serialization struct {
	Node struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"node"`
}
