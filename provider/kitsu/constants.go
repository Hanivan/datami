package kitsu

const (
	OpGet         = "get"
	OpSearch      = "search"
	OpGetByMalID  = "getbymalid"
	OpGetManga    = "getmanga"
	OpSearchManga = "searchmanga"
)

const (
	defaultBaseURL = "https://kitsu.io/api/edge"
	defaultRate    = 4
)

var kitsuStatusMap = map[string]string{
	"current":    "ongoing",
	"finished":   "finished",
	"tba":        "planned",
	"upcoming":   "planned",
	"unreleased": "planned",
}

var kitsuShowTypeMap = map[string]string{
	"TV":      "tv",
	"special": "special",
	"OVA":     "ova",
	"ONA":     "ona",
	"movie":   "movie",
	"music":   "music",
}
