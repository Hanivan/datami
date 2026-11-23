package helpers

import (
	"errors"
	"testing"

	"github.com/Hanivan/datami/types"
)

func TestWrap(t *testing.T) {
	err := Wrap(types.SourceMAL, "get", types.ErrNotFound)
	var te *types.Error
	if !errors.As(err, &te) || te.Source != types.SourceMAL || te.Op != "get" || !errors.Is(err, types.ErrNotFound) {
		t.Fatalf("got %#v", err)
	}
}
