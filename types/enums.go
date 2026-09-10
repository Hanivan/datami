package types

// MediaStatus represents the release status of a media entry.
type MediaStatus string

const (
	StatusFinished     MediaStatus = "finished"
	StatusOngoing      MediaStatus = "ongoing"
	StatusPlanned      MediaStatus = "planned"
	StatusCancelled    MediaStatus = "cancelled"
	StatusHiatus       MediaStatus = "hiatus"
	StatusDiscontinued MediaStatus = "discontinued"
)

// MediaType represents the format/type of a media entry.
type MediaType string

const (
	TypeTV         MediaType = "tv"
	TypeTVShort    MediaType = "tv_short"
	TypeMovie      MediaType = "movie"
	TypeSpecial    MediaType = "special"
	TypeOVA        MediaType = "ova"
	TypeONA        MediaType = "ona"
	TypeMusic      MediaType = "music"
	TypeManga      MediaType = "manga"
	TypeNovel      MediaType = "novel"
	TypeOneShot    MediaType = "one_shot"
	TypeDoujinshi  MediaType = "doujinshi"
	TypeManhwa     MediaType = "manhwa"
	TypeManhua     MediaType = "manhua"
	TypeOEL        MediaType = "oel"
	TypeLightNovel MediaType = "light_novel"
)

// TitleKind represents the kind/variant of a title.
type TitleKind string

const (
	KindCanonical TitleKind = "canonical"
	KindRomaji    TitleKind = "romaji"
	KindEnglish   TitleKind = "english"
	KindNative    TitleKind = "native"
	KindSynonym   TitleKind = "synonym"
	KindShort     TitleKind = "short"
)

// Season represents a release season.
type Season string

const (
	SeasonWinter Season = "winter"
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonFall   Season = "fall"
)

// RelationKind represents the type of relation between media entries.
type RelationKind string

const (
	RelationSequel      RelationKind = "sequel"
	RelationPrequel     RelationKind = "prequel"
	RelationSideStory   RelationKind = "side_story"
	RelationParentStory RelationKind = "parent_story"
	RelationSummary     RelationKind = "summary"
	RelationAlternative RelationKind = "alternative"
	RelationSpinOff     RelationKind = "spin_off"
	RelationAdaptation  RelationKind = "adaptation"
	RelationCharacter   RelationKind = "character"
	RelationOther       RelationKind = "other"
	RelationSource      RelationKind = "source"
	RelationCompilation RelationKind = "compilation"
	RelationContains    RelationKind = "contains"
)

// ImageSize represents the size variant of an image.
type ImageSize string

const (
	SizeMedium   ImageSize = "medium"
	SizeLarge    ImageSize = "large"
	SizeOriginal ImageSize = "original"
)

// ResourceType represents the type of an external resource link.
type ResourceType string

const (
	ResourceAnime ResourceType = "anime"
	ResourceManga ResourceType = "manga"
)

// CharacterRoleType represents the role type of a character in a media.
type CharacterRoleType string

const (
	RoleMain       CharacterRoleType = "main"
	RoleSupporting CharacterRoleType = "supporting"
)
