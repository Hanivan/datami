package mal

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
