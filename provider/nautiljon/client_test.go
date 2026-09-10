package nautiljon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Hanivan/datami/types"
)

const nautiljonListHTML = `<!DOCTYPE html><html><body>
<table class="search"><tbody>
<tr><td></td><td>Cowboy Bebop</td><td>TV</td><td>terminé</td></tr>
<tr><td></td><td>Monster</td><td>Manga</td><td>en cours</td></tr>
</tbody></table>
</body></html>`

func TestSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(nautiljonListHTML))
	}))
	defer srv.Close()

	c := NewWithTestServer(srv.URL, Config{Timeout: 5 * time.Second})
	results, err := c.Search(context.Background(), "", types.SearchOpts{Page: 1})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Titles[0].Text != "Cowboy Bebop" {
		t.Fatalf("expected title Cowboy Bebop, got %s", results[0].Titles[0].Text)
	}
	if results[0].Status != "finished" {
		t.Fatalf("expected status finished, got %s", results[0].Status)
	}
}

func TestGetNotImplemented(t *testing.T) {
	c := NewWithTestServer("http://dummy", Config{Timeout: 5 * time.Second})
	_, err := c.Get(context.Background(), "1")
	if err == nil {
		t.Fatal("expected error for Get")
	}
}
