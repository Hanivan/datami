package types

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrRateLimited    = errors.New("rate limited")
	ErrAuthFailed     = errors.New("authentication failed")
	ErrUnavailable    = errors.New("source unavailable")
	ErrInvalidInput   = errors.New("invalid input")
	ErrNotImplemented = errors.New("not implemented")
)

type Error struct {
	Source string
	Op     string
	Err    error
	// Status is the HTTP status the provider answered with, 0 when the failure
	// was not an HTTP response. Callers deciding whether to back off must branch
	// on this, never on the message text: providers word the same condition
	// differently (AniList alone uses two distinct 403 bodies) and a reworded
	// string would silently disable the backoff.
	Status int
	// RetryAfter is the server's own requested cooldown in seconds, taken from
	// the Retry-After header. 0 when the response carried none.
	RetryAfter int
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Source + "." + e.Op + ": " + e.Err.Error()
	}
	return e.Source + "." + e.Op
}

func (e *Error) Unwrap() error {
	return e.Err
}

// StatusOf returns the HTTP status carried by err, or 0 when err is not (and
// does not wrap) an *Error with a status. It is the supported way to branch on
// a provider failure's status.
func StatusOf(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Status
	}
	return 0
}

// RetryAfterOf returns the server-requested cooldown in seconds carried by err,
// or 0 when there is none.
func RetryAfterOf(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.RetryAfter
	}
	return 0
}
