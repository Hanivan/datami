package httpx

import (
	"errors"
	"net/http"
	"time"

	"github.com/Hanivan/datami/types"
)

const (
	DefaultTimeout      = 10 * time.Second
	MaxRetries          = 3
	RetryBackoff        = 100 * time.Millisecond
	DefaultSearchLimit  = 10
	ServerErrThreshold  = 500
	ContentTypeJSON     = "application/json"
	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
	BearerPrefix        = "Bearer "
	// DefaultUserAgent names datami when the caller sets none, so no request
	// goes out as Go's anonymous "Go-http-client/1.1".
	DefaultUserAgent = "datami (+https://github.com/Hanivan/datami)"
	// DefaultMaxBody caps a response body read into memory.
	DefaultMaxBody = 32 << 20
)

// ErrBodyTooLarge marks a response body over Config.MaxBody. Not retried.
var ErrBodyTooLarge = errors.New("httpx: response body too large")

// ErrLimiterWait marks a failed rate limiter wait: no request was sent. The
// limiter or ctx error stays in the chain.
var ErrLimiterWait = errors.New("httpx: rate limiter wait")

type Config struct {
	Timeout time.Duration
	Rate    float64
	Burst   int
	// Limiter, when non-nil, replaces the limiter built from Rate/Burst so
	// callers can share one budget across clients.
	Limiter   types.Limiter
	Metrics   func(op string, dur time.Duration, err error)
	Transport http.RoundTripper
	Headers   map[string]string
	// UserAgent is sent on every request; "" uses DefaultUserAgent. Name your
	// app here so upstreams can identify and throttle it.
	UserAgent string
	// MaxRetries is total attempts per request; 0 uses the MaxRetries const.
	MaxRetries int
	// MaxBody caps the response body in bytes; 0 uses DefaultMaxBody.
	MaxBody int64
}
