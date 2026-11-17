package helpers

import (
	"errors"
	"strings"
	"testing"

	"github.com/Hanivan/datami/types"
)

func TestStatusError(t *testing.T) {
	cases := []struct {
		code int
		want error
	}{
		{404, types.ErrNotFound},
		{429, types.ErrRateLimited},
		{401, types.ErrAuthFailed},
		{403, types.ErrAuthFailed},
	}
	for _, c := range cases {
		err := StatusError(types.SourceMAL, "get", c.code, 0)
		if !errors.Is(err, c.want) {
			t.Errorf("%d: want %v got %v", c.code, c.want, err)
		}
		if got := types.StatusOf(err); got != c.code {
			t.Errorf("%d: status %d", c.code, got)
		}
	}
	for _, src := range []string{types.SourceHikka, types.SourceKaize, types.SourceNautiljon, types.SourceOtakotaku} {
		if err := StatusError(src, "get", 403, 0); !errors.Is(err, types.ErrRateLimited) {
			t.Errorf("%s 403: want ErrRateLimited got %v", src, err)
		}
		if err := StatusError(src, "get", 401, 0); !errors.Is(err, types.ErrAuthFailed) {
			t.Errorf("%s 401: want ErrAuthFailed got %v", src, err)
		}
	}
	err := StatusError(types.SourceMAL, "get", 500, 0)
	if types.StatusOf(err) != 500 || !strings.Contains(err.Error(), "status 500") {
		t.Errorf("500: %v", err)
	}
}

func TestDecodeJSON(t *testing.T) {
	v, err := DecodeJSON[struct{ A int }](types.SourceMAL, "get", 200, 0, []byte(`{"A":1}`))
	if err != nil || v.A != 1 {
		t.Fatalf("got %v %v", v, err)
	}
	_, err = DecodeJSON[struct{ A int }](types.SourceMAL, "get", 200, 0, []byte(`{`))
	if err == nil || !strings.Contains(err.Error(), "parse:") {
		t.Fatalf("got %v", err)
	}
	_, err = DecodeJSON[struct{ A int }](types.SourceMAL, "get", 404, 0, nil)
	if !errors.Is(err, types.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestRetryAfterCarried(t *testing.T) {
	if got := types.RetryAfterOf(StatusError(types.SourceMAL, "get", 429, 7)); got != 7 {
		t.Fatalf("StatusError retryAfter = %d, want 7", got)
	}
	_, err := DecodeJSON[struct{ A int }](types.SourceMAL, "get", 503, 9, nil)
	if got := types.RetryAfterOf(err); got != 9 {
		t.Fatalf("DecodeJSON retryAfter = %d, want 9", got)
	}
}
