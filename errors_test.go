package datami

import (
	"errors"
	"testing"

	"github.com/Hanivan/datami/types"
)

func TestErrorUnwrap(t *testing.T) {
	inner := errors.New("inner")
	e := &Error{Source: types.SourceMAL, Op: "get", Err: inner}
	if !errors.Is(e, inner) {
		t.Fatal("errors.Is should match inner error")
	}
	if e.Source != types.SourceMAL {
		t.Fatalf("expected source mal, got %s", e.Source)
	}
	if e.Op != "get" {
		t.Fatalf("expected op get, got %s", e.Op)
	}
}

func TestSentinelErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"NotFound", ErrNotFound},
		{"RateLimited", ErrRateLimited},
		{"AuthFailed", ErrAuthFailed},
		{"Unavailable", ErrUnavailable},
		{"InvalidInput", ErrInvalidInput},
	}
	for _, c := range cases {
		if c.err == nil {
			t.Fatalf("%s is nil", c.name)
		}
		if c.err.Error() == "" {
			t.Fatalf("%s has empty message", c.name)
		}
	}
}
