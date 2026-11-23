package mal

import (
	"time"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

type Config struct {
	ClientID string
	Timeout  time.Duration
	// UserAgent names the calling app on every request; "" sends
	// httpx.DefaultUserAgent.
	UserAgent string
	// Limiter replaces the built-in rate limiter when non-nil, so callers can
	// share one request budget across clients.
	Limiter types.Limiter
	Metrics func(source string, op string, dur time.Duration, err error)
}

type Client struct {
	baseURL string
	client  *httpx.Client
	metrics func(source string, op string, dur time.Duration, err error)
}

func New(cfg Config) (*Client, error) {
	return newClient("https://api.myanimelist.net/v2", cfg)
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
			UserAgent: cfg.UserAgent,
			Limiter:   cfg.Limiter,
			Rate:      2,
			Headers:   map[string]string{"X-MAL-CLIENT-ID": cfg.ClientID},
			Metrics:   httpx.SourceMetrics(types.SourceMAL, cfg.Metrics),
		}),
		metrics: cfg.Metrics,
	}, nil
}
