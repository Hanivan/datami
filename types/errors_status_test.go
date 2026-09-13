package types

import (
	"errors"
	"net/http"
	"testing"
)

// TestErrorCarriesStatus pins the status contract. A consumer deciding whether
// to back off must branch on the HTTP status, never on the message text: the
// same condition is worded differently across providers and rewording a string
// would silently disable the backoff.
func TestErrorCarriesStatus(t *testing.T) {
	e := &Error{Source: SourceAniList, Op: "get", Status: http.StatusTooManyRequests, Err: ErrRateLimited}
	if e.Status != http.StatusTooManyRequests {
		t.Fatalf("status not carried: %d", e.Status)
	}
	if !errors.Is(e, ErrRateLimited) {
		t.Fatal("sentinel must still unwrap")
	}
}

// TestStatusOfReadsWrappedError asserts StatusOf digs the status out of a
// wrapped *Error, which is how callers receive it.
func TestStatusOfReadsWrappedError(t *testing.T) {
	wrapped := errors.Join(errors.New("context"), &Error{Source: SourceAniList, Op: "get", Status: 403})
	if got := StatusOf(wrapped); got != 403 {
		t.Fatalf("StatusOf = %d, want 403", got)
	}
	if got := StatusOf(errors.New("plain")); got != 0 {
		t.Fatalf("StatusOf on a non-status error = %d, want 0", got)
	}
}

// TestRetryAfterRoundTrips asserts the server's own cooldown survives, so a
// caller can honour it instead of guessing a backoff.
func TestRetryAfterRoundTrips(t *testing.T) {
	e := &Error{Source: SourceAniList, Op: "get", Status: 429, RetryAfter: 30}
	if e.RetryAfter != 30 {
		t.Fatalf("retry-after not carried: %d", e.RetryAfter)
	}
}
