package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

// DecodeJSON returns StatusError for non-200 codes, else unmarshals body into T.
func DecodeJSON[T any](source, op string, code, retryAfter int, body []byte) (T, error) {
	var v T
	if code != http.StatusOK {
		return v, StatusError(source, op, code, retryAfter)
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return v, &types.Error{Source: source, Op: op, Err: fmt.Errorf("parse: %w", err)}
	}
	return v, nil
}

// FetchJSON GETs url through c and decodes the JSON body into T.
func FetchJSON[T any](ctx context.Context, c *httpx.Client, source, op, url string) (T, error) {
	code, body, retryAfter, err := c.Get(ctx, op, url)
	if err != nil {
		var zero T
		return zero, Wrap(source, op, err)
	}
	return DecodeJSON[T](source, op, code, retryAfter, body)
}
