package mal

const (
	OpGet                   = "get"
	OpSearch                = "search"
	OpGetManga              = "getmanga"
	OpSearchManga           = "searchmanga"
	OpGetSeasonal           = "getseasonal"
	OpGetRanking            = "getranking"
	OpGetMangaRanking       = "getmangaranking"
	OpGetAnimeCharacters    = "getanimecharacters"
	OpGetCharacter          = "getcharacter"
	OpGetPerson             = "getperson"
	OpGetForumBoards        = "getforumboards"
	OpGetForumTopic         = "getforumtopic"
	OpSearchForumTopics     = "searchforumtopics"
	OpGetUserAnimeList      = "getuseranimelist"
	OpGetUserMangaList      = "getusermangalist"
	OpGetUserInfo           = "getuserinfo"
	OpUpdateAnimeListStatus = "updateanimeliststatus"
	OpDeleteAnimeListItem   = "deleteanimelistitem"
)

// MAL's list endpoints (search, seasonal, ranking) return only id, title and
// main_picture unless `fields` asks for more, unlike the by-id endpoints,
// which fail loudly on a bad field. A list request for an unsupported field is
// silently dropped from the response, so anime and manga need separate sets:
// asking a manga endpoint for num_episodes yields no error and no value.
const (
	animeListFields = "id,title,main_picture,media_type,status,num_episodes,mean"
	mangaListFields = "id,title,main_picture,media_type,status,num_chapters,num_volumes,mean"
)

var malStatusMap = map[string]string{
	"finished_airing":      "finished",
	"currently_airing":     "ongoing",
	"not_yet_aired":        "planned",
	"finished_publishing":  "finished",
	"currently_publishing": "ongoing",
	"not_yet_published":    "planned",
	"on_hiatus":            "hiatus",
	"discontinued":         "discontinued",
}

var malMediaTypeMap = map[string]string{
	"tv":          "tv",
	"ova":         "ova",
	"movie":       "movie",
	"special":     "special",
	"ona":         "ona",
	"music":       "music",
	"manga":       "manga",
	"novel":       "novel",
	"one_shot":    "one_shot",
	"doujinshi":   "doujinshi",
	"manhwa":      "manhwa",
	"manhua":      "manhua",
	"oel":         "oel",
	"light_novel": "light_novel",
}
