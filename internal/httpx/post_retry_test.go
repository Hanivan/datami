package httpx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestPostRetryResendsBody(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if atomic.AddInt32(&hits, 1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Write(b)
	}))
	defer srv.Close()

	code, body, _, err := New(Config{}).Post(context.Background(), "op", srv.URL, []byte(`{"q":1}`), nil)
	if err != nil || code != 200 || string(body) != `{"q":1}` || hits != 2 {
		t.Fatalf("code %d body %q err %v hits %d", code, body, err, hits)
	}
}
