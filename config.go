package datami

import (
	"time"

	"github.com/Hanivan/datami/types"
)

type (
	Media       = types.Media
	Title       = types.Title
	Date        = types.Date
	Score       = types.Score
	Studio      = types.Studio
	Image       = types.Image
	ExternalID  = types.ExternalID
	Relation    = types.Relation
	SearchOpts  = types.SearchOpts
	Error       = types.Error
	MetricsHook = types.MetricsHook
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
	MALClientID   string
	AniListToken  string
	Timeout       time.Duration
	RateLimitRate float64
	Metrics       MetricsHook
}
