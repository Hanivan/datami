package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultTimeout      = 10 * time.Second
	MaxRetries          = 3
	RetryBackoff        = 100 * time.Millisecond
	DefaultSearchLimit  = 10
	ServerErrThreshold  = 500
	ContentTypeJSON     = "application/json"
	HeaderContentType   = "Content-Type"
	HeaderAccept        = "Accept"
	HeaderAuthorization = "Authorization"
	BearerPrefix        = "Bearer "
)

type Config struct {
	Timeout   time.Duration
	Rate      float64
	Burst     int
	Metrics   func(op string, dur time.Duration, err error)
	Transport http.RoundTripper
}

type Client struct {
	base    *http.Client
	limiter *rate.Limiter
	metrics func(op string, dur time.Duration, err error)
}

func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	burst := cfg.Burst
	if burst == 0 {
		burst = 1
	}
	var limiter *rate.Limiter
	if cfg.Rate > 0 {
		limiter = rate.NewLimiter(rate.Limit(cfg.Rate), burst)
	}
	base := &http.Client{
		Timeout: timeout,
	}
	if cfg.Transport != nil {
		base.Transport = cfg.Transport
	}
	return &Client{
		base:    base,
		limiter: limiter,
		metrics: cfg.Metrics,
	}
}

func (c *Client) Get(ctx context.Context, url string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}
	return c.Do(req)
}

func (c *Client) Post(ctx context.Context, url string, body []byte, headers map[string]string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (int, []byte, error) {
	if c.limiter != nil {
		if err := c.limiter.Wait(req.Context()); err != nil {
			return 0, nil, fmt.Errorf("rate limiter: %w", err)
		}
	}
	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * RetryBackoff)
		}
		start := time.Now()
		resp, err := c.base.Do(req)
		if err != nil {
			lastErr = err
			if c.metrics != nil {
				c.metrics(req.Method, time.Since(start), err)
			}
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if c.metrics != nil {
			c.metrics(req.Method, time.Since(start), readErr)
		}
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= ServerErrThreshold {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}
		return resp.StatusCode, body, nil
	}
	return 0, nil, fmt.Errorf("after %d attempts: %w", MaxRetries, lastErr)
}
