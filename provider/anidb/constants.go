package anidb

import (
	"fmt"
	"time"

	"github.com/Hanivan/datami/types"
)

const (
	OpNew             = "new"
	OpGet             = "get"
	OpSearch          = "search"
	OpGetByMalID      = "getbymalid"
	OpGetManga        = "getmanga"
	OpSearchManga     = "searchmanga"
	OpGetMangaByMalID = "getmangabymalid"
	OpGetCharacters   = "getcharacters"
)

const (
	// defaultAPIURL is plain HTTP: AniDB serves no TLS on port 9001.
	defaultAPIURL    = "http://api.anidb.net:9001/httpapi"
	defaultTitlesURL = "https://anidb.net/api/anime-titles.xml.gz"
)

const (
	titlesMaxAge        = 24 * time.Hour
	titlesRetryCooldown = time.Hour
	titlesFile          = "anime-titles.xml.gz"
)

const (
	// maxAnimeXML caps a decompressed anime response.
	maxAnimeXML = 16 << 20
	// maxTitlesDump caps the decompressed titles dump.
	maxTitlesDump = 256 << 20
)

var errTooLarge = fmt.Errorf("%w: response too large", types.ErrUnavailable)

const (
	imageBaseURL = "https://cdn.anidb.net/images/main/"
	// genreMinWeight is AniDB's genre-like tag threshold; all tags stay in Raw.
	genreMinWeight = 400
	// Resource types from the <resources> block.
	resourceANN = 1
	resourceMAL = 2
	// characterTypeCharacter is charactertype id 1; others are organizations, vessels, etc.
	characterTypeCharacter = 1
)

var anidbTypeMap = map[string]string{
	"TV Series":   "tv",
	"TV Special":  "special",
	"Movie":       "movie",
	"OVA":         "ova",
	"ONA":         "ona",
	"Music Video": "music",
	"Web":         "ona",
	"Manga":       "manga",
	"Novel":       "novel",
	"Other":       "other",
	"Unknown":     "unknown",
}

var relationMap = map[string]types.RelationKind{
	"Sequel":              types.RelationSequel,
	"Prequel":             types.RelationPrequel,
	"Side Story":          types.RelationSideStory,
	"Parent Story":        types.RelationParentStory,
	"Summary":             types.RelationSummary,
	"Full Story":          types.RelationParentStory,
	"Alternative Setting": types.RelationAlternative,
	"Alternative Version": types.RelationAlternative,
	"Character":           types.RelationCharacter,
	"Same Setting":        types.RelationOther,
	"Other":               types.RelationOther,
}

var titleKindMap = map[string]types.TitleKind{
	"en":    types.KindEnglish,
	"ja":    types.KindNative,
	"x-jat": types.KindRomaji,
}

var resourceSourceMap = map[int]string{
	resourceANN: types.SourceAnimeNewsNetwork,
	resourceMAL: types.SourceMAL,
}
