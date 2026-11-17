package hikka

import (
	"time"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
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
	return newClient("https://api.hikka.io", cfg)
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
			Rate:    2,
			Metrics: func(op string, dur time.Duration, err error) {
				if cfg.Metrics != nil {
					cfg.Metrics(types.SourceHikka, op, dur, err)
				}
			},
		}),
	}, nil
}
