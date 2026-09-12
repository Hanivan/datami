package mal

import (
	"strconv"
	"strings"
)

type anime struct {
	ID                int            `json:"id"`
	Title             string         `json:"title"`
	MainPicture       *picture       `json:"main_picture,omitempty"`
	Synopsis          string         `json:"synopsis,omitempty"`
	StartDate         string         `json:"start_date,omitempty"`
	EndDate           string         `json:"end_date,omitempty"`
	Status            string         `json:"status,omitempty"`
	MediaType         string         `json:"media_type,omitempty"`
	NumEpisodes       int            `json:"num_episodes,omitempty"`
	AverageEpDuration int            `json:"average_episode_duration,omitempty"`
	Mean              float64        `json:"mean,omitempty"`
	Rank              int            `json:"rank,omitempty"`
	Popularity        int            `json:"popularity,omitempty"`
	NumListUsers      int            `json:"num_list_users,omitempty"`
	NumScoringUsers   int            `json:"num_scoring_users,omitempty"`
	NumFavorites      int            `json:"num_favorites,omitempty"`
	NSFW              string         `json:"nsfw,omitempty"`
	Genres            []genre        `json:"genres,omitempty"`
	Studios           []studio       `json:"studios,omitempty"`
	AlternativeTitles *altTitles     `json:"alternative_titles,omitempty"`
	RelatedAnime      []relatedEntry `json:"related_anime,omitempty"`
	RelatedManga      []relatedEntry `json:"related_manga,omitempty"`
	StartSeason       struct {
		Season string `json:"season"`
		Year   int    `json:"year"`
	} `json:"start_season,omitempty"`
	Broadcast       *broadcast       `json:"broadcast,omitempty"`
	Source          string           `json:"source,omitempty"`
	Rating          string           `json:"rating,omitempty"`
	Pictures        []picture        `json:"pictures,omitempty"`
	Background      string           `json:"background,omitempty"`
	Recommendations []recommendation `json:"recommendations,omitempty"`
	Statistics      *statistics      `json:"statistics,omitempty"`
	MyListStatus    *listStatus      `json:"my_list_status,omitempty"`
	CreatedAt       string           `json:"created_at,omitempty"`
	UpdatedAt       string           `json:"updated_at,omitempty"`
}

type broadcast struct {
	DayOfTheWeek string `json:"day_of_the_week"`
	StartTime    string `json:"start_time"`
}

type recommendation struct {
	Node               animeRelation `json:"node"`
	NumRecommendations int           `json:"num_recommendations"`
}

type statistics struct {
	Status struct {
		Watching    flexInt `json:"watching"`
		Completed   flexInt `json:"completed"`
		OnHold      flexInt `json:"on_hold"`
		Dropped     flexInt `json:"dropped"`
		PlanToWatch flexInt `json:"plan_to_watch"`
	} `json:"status"`
	NumListUsers flexInt `json:"num_list_users"`
}

type flexInt int64

func (f *flexInt) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	s := strings.Trim(string(b), `"`)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*f = flexInt(n)
	return nil
}

type listStatus struct {
	Status             string `json:"status"`
	Score              int    `json:"score"`
	NumWatchedEpisodes int    `json:"num_watched_episodes"`
	IsRewatching       bool   `json:"is_rewatching"`
}

type picture struct {
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type studio struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type altTitles struct {
	EN       string   `json:"en"`
	JA       string   `json:"ja"`
	Synonyms []string `json:"synonyms"`
}

type relatedEntry struct {
	Node         animeRelation `json:"node"`
	RelationType string        `json:"relation_type"`
}

type animeRelation struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	// MediaType is MAL's verbatim node media_type ("tv"/"ova"/"movie"/"manga"/
	// "light_novel"/...). Only returned when the request asks for it via the
	// nested field selector related_anime{media_type}; consumers classify the
	// relation target from it, so a bare related_anime request makes every
	// target indistinguishable.
	MediaType string `json:"media_type,omitempty"`
}

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

type character struct {
	ID              int      `json:"id"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	AlternativeName string   `json:"alternative_name,omitempty"`
	MainPicture     *picture `json:"main_picture,omitempty"`
	Biography       string   `json:"biography,omitempty"`
	// NumFavorites is a pointer: 0 is a real favourite count and must not be
	// indistinguishable from a field MAL did not return.
	NumFavorites *int `json:"num_favorites,omitempty"`
}

type animeCharacter struct {
	Node character `json:"node"`
	Role string    `json:"role"`
}

type person struct {
	ID           int      `json:"id"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	MainPicture  *picture `json:"main_picture,omitempty"`
	Birthday     string   `json:"birthday,omitempty"`
	NumFavorites *int     `json:"num_favorites,omitempty"`
	More         string   `json:"more,omitempty"`
}

type searchResponse struct {
	Data   []searchNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type searchNode struct {
	Node anime `json:"node"`
}

type mangaSearchResponse struct {
	Data   []mangaSearchNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type mangaSearchNode struct {
	Node manga `json:"node"`
}

type seasonalResponse struct {
	Data   []searchNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type rankingResponse struct {
	Data   []rankingNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type rankingNode struct {
	Node    anime `json:"node"`
	Ranking struct {
		Rank int `json:"rank"`
	} `json:"ranking"`
}

type charactersResponse struct {
	Data   []animeCharacter `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}
