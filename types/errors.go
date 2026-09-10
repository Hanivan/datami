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
