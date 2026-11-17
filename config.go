package datami

import (
	"time"

	"github.com/Hanivan/datami/types"
)

type (
	Media        = types.Media
	Title        = types.Title
	Date         = types.Date
	Score        = types.Score
	Studio       = types.Studio
	Image        = types.Image
	ExternalID   = types.ExternalID
	Relation     = types.Relation
	SearchOpts   = types.SearchOpts
	Error        = types.Error
	MetricsHook  = types.MetricsHook
	TagsLinks    = types.TagsLinks
	Tag          = types.Tag
	ExternalLink = types.ExternalLink
)

// Re-export sentinel errors
var (
	ErrNotFound     = types.ErrNotFound
	ErrRateLimited  = types.ErrRateLimited
	ErrAuthFailed   = types.ErrAuthFailed
	ErrUnavailable  = types.ErrUnavailable
	ErrInvalidInput = types.ErrInvalidInput
)

type Config struct {
	MALClientID  string
	AniListToken string
	// AniListDescriptionAsHTML: see anilist.Config.DescriptionAsHTML.
	AniListDescriptionAsHTML *bool
	// AniDB registered HTTP client. Empty AniDBClientName leaves AniDB
	// unconfigured: its calls return ErrInvalidInput.
	AniDBClientName    string
	AniDBClientVersion int
	AniDBCacheDir      string
	Timeout            time.Duration
	Metrics            MetricsHook
}
