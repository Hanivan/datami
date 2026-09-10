package anilist

import (
	"time"

	"github.com/Hanivan/datami/internal/httpx"
)

type Config struct {
	Token   string
	Timeout time.Duration
	Metrics func(source string, op string, dur time.Duration, err error)
}

type Client struct {
	baseURL string
	client  *httpx.Client
	token   string
	metrics func(source string, op string, dur time.Duration, err error)
}

func New(cfg Config) (*Client, error) {
	return newClient("https://graphql.anilist.co", cfg)
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
			Rate:    1.5,
			Metrics: func(op string, dur time.Duration, err error) {
				if cfg.Metrics != nil {
					cfg.Metrics("anilist", op, dur, err)
				}
			},
		}),
		token:   cfg.Token,
		metrics: cfg.Metrics,
	}, nil
}
