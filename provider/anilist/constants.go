package anilist

import "errors"

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

const (
	OpGet                    = "get"
	OpSearch                 = "search"
	OpGetByMalID             = "getbymalid"
	OpGetManga               = "getmanga"
	OpSearchManga            = "searchmanga"
	OpGetMangaByMalID        = "getmangabymalid"
	OpGetByMalIDs            = "getbymalids"
	OpGetMangaByMalIDs       = "getmangabymalids"
	OpGetViewer              = "getviewer"
	OpGetMediaListEntry      = "getmedialistentry"
	OpGetMediaListCollection = "getmedialistcollection"
	OpSaveMediaListEntry     = "savemedialistentry"
	OpDeleteMediaListEntry   = "deletemedialistentry"
	OpGetCharacters          = "getcharacters"
	OpGetStaff               = "getstaff"
	OpFollowUps              = "followups"
	OpGetTagsLinks           = "gettagslinks"
)

// followUpAliases caps the aliased connections packed into one request.
const followUpAliases = 8

// maxBatch is AniList's perPage ceiling.
const maxBatch = 50

const (
	// maxPages is the default total page cap per connection (embedded +
	// follow-up), and the fixed cap of the standalone GetCharacters/GetStaff walk.
	maxPages = 10
	// pageSize is AniList's per-page ceiling; never send more.
	pageSize = 25

	charSort  = "sort: ROLE"
	staffSort = "sort: RELEVANCE"

	// staffBody is the selection inside a staff connection, shared by the
	// standalone and embedded queries.
	staffBody = `{ pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } primaryOccupations } } }`
)

// ErrMediaNotFound is returned by the tags/links fetches when AniList answers
// a resolved, error-free {"data":{"Media":null}} - the id did not resolve to a
// Media at all. Deliberately distinct from a nil error with empty slices,
// which means the Media DID resolve and genuinely has none: a caller doing a
// delete-then-insert refresh must tell the two apart, since "didn't resolve
// this time" is not grounds to conclude the prior set is stale.
var ErrMediaNotFound = errors.New("anilist: media not found")
