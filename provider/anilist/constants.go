package anilist

var anilistStatusMap = map[string]string{
	"FINISHED":         "finished",
	"RELEASING":        "ongoing",
	"NOT_YET_RELEASED": "planned",
	"CANCELLED":        "cancelled",
	"HIATUS":           "hiatus",
}

var anilistFormatMap = map[string]string{
	"TV":       "tv",
	"TV_SHORT": "tv_short",
	"MOVIE":    "movie",
	"SPECIAL":  "special",
	"OVA":      "ova",
	"ONA":      "ona",
	"MUSIC":    "music",
	"MANGA":    "manga",
	"NOVEL":    "novel",
	"ONE_SHOT": "one_shot",
}
