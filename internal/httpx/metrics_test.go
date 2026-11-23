package httpx

import (
	"testing"
	"time"
)

func TestSourceMetricsNil(t *testing.T) {
	if SourceMetrics("x", nil) != nil {
		t.Fatal("want nil")
	}
}

func TestSourceMetricsInjects(t *testing.T) {
	var gs, go_ string
	f := SourceMetrics("x", func(s, o string, _ time.Duration, _ error) { gs, go_ = s, o })
	f("get", 0, nil)
	if gs != "x" || go_ != "get" {
		t.Fatalf("got %q %q", gs, go_)
	}
}
