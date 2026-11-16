package anilist

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestDescriptionAsHTML pins that Config.DescriptionAsHTML reaches every
// description selection, and that nil defaults to HTML.
func TestDescriptionAsHTML(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name string
		v    *bool
		want string
	}{
		{"default", nil, "description(asHtml: true)"},
		{"html", &yes, "description(asHtml: true)"},
		{"markdown", &no, "description(asHtml: false)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var bodies []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				bodies = append(bodies, string(b))
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"data":{"Media":{"id":1,"characters":{"pageInfo":{"hasNextPage":false},"edges":[]}}}}`))
			}))
			defer srv.Close()
			c := NewWithTestServer(srv.URL, Config{DescriptionAsHTML: tc.v})
			c.Get(context.Background(), "1")
			c.GetCharacters(context.Background(), "1")
			for _, b := range bodies {
				if !strings.Contains(b, tc.want) {
					t.Fatalf("want %q in query, got %s", tc.want, b)
				}
			}
			if len(bodies) != 2 {
				t.Fatalf("want 2 requests, got %d", len(bodies))
			}
		})
	}
}
