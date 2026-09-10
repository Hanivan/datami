package kitsu

import (
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
	metrics func(source string, op string, dur time.Duration, err error)
}

func New(cfg Config) (*Client, error) {
	return newClient("https://kitsu.io/api/edge", cfg)
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
			Timeout: timeout,
			Rate:    4,
			Metrics: func(op string, dur time.Duration, err error) {
				if cfg.Metrics != nil {
					cfg.Metrics("kitsu", op, dur, err)
				}
			},
		}),
		metrics: cfg.Metrics,
	}, nil
}
