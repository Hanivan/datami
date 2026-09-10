package nautiljon

import (
	"net/http"
	"time"

	"github.com/linkeunid/datami/internal/httpx"
)

type Config struct {
	Timeout time.Duration
	Metrics func(source string, op string, dur time.Duration, err error)
}

type Client struct {
	baseURL string
	client  *httpx.Client
}

func New(cfg Config) (*Client, error) {
	return newClient("https://www.nautiljon.com", cfg)
}

func NewWithTestServer(baseURL string, cfg Config) *Client {
	c, _ := newClient(baseURL, cfg)
	return c
}

func newClient(baseURL string, cfg Config) (*Client, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL: baseURL,
		client: httpx.New(httpx.Config{
			Timeout:   timeout,
			Rate:      1,
			Transport: &rewriteTransport{base: baseURL},
			Metrics: func(op string, dur time.Duration, err error) {
				if cfg.Metrics != nil {
					cfg.Metrics("nautiljon", op, dur, err)
				}
			},
		}),
	}, nil
}

type rewriteTransport struct {
	base string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// no-op; real impl would use cloudscraper. kept for parity.
	return http.DefaultTransport.RoundTrip(req)
}
