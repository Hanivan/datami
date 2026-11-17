package anidb

import "github.com/Hanivan/datami/types"

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
