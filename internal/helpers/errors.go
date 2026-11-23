package helpers

import (
	"fmt"
	"net/http"

	"github.com/Hanivan/datami/types"
)

// keyless sources send no credentials, so a 403 from them is a bot/WAF block
// (back off and retry), not a credential problem.
var keyless = map[string]bool{
	types.SourceHikka:     true,
	types.SourceKaize:     true,
	types.SourceNautiljon: true,
	types.SourceOtakotaku: true,
}

// StatusError maps a non-200 HTTP status to a *types.Error carrying the status.
func StatusError(source, op string, code, retryAfter int) error {
	var err error
	switch code {
	case http.StatusNotFound:
		err = types.ErrNotFound
	case http.StatusTooManyRequests:
		err = types.ErrRateLimited
	case http.StatusForbidden:
		err = types.ErrAuthFailed
		if keyless[source] {
			err = types.ErrRateLimited
		}
	case http.StatusUnauthorized:
		err = types.ErrAuthFailed
	default:
		err = fmt.Errorf("status %d", code)
	}
	return &types.Error{Source: source, Op: op, Status: code, RetryAfter: retryAfter, Err: err}
}

// Wrap tags err with the source and operation it came from.
func Wrap(source, op string, err error) error {
	return &types.Error{Source: source, Op: op, Err: err}
}
