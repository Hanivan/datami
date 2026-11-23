package httpx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/Hanivan/datami/types"
)

type Client struct {
	base       *http.Client
	limiter    types.Limiter
	metrics    func(op string, dur time.Duration, err error)
	headers    map[string]string
	userAgent  string
	maxRetries int
	maxBody    int64
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
	limiter := cfg.Limiter
	if limiter == nil && cfg.Rate > 0 {
		limiter = rate.NewLimiter(rate.Limit(cfg.Rate), burst)
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = MaxRetries
	}
	maxBody := cfg.MaxBody
	if maxBody <= 0 {
		maxBody = DefaultMaxBody
	}
	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}
	base := &http.Client{
		Timeout: timeout,
	}
	if cfg.Transport != nil {
		base.Transport = cfg.Transport
	}
	return &Client{
		base:       base,
		limiter:    limiter,
		metrics:    cfg.Metrics,
		headers:    cfg.Headers,
		userAgent:  userAgent,
		maxRetries: maxRetries,
		maxBody:    maxBody,
	}
}

func (c *Client) Get(ctx context.Context, op, url string) (int, []byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, 0, fmt.Errorf("new request: %w", err)
	}
	return c.Do(op, req)
}

func (c *Client) Post(ctx context.Context, op, url string, body []byte, headers map[string]string) (int, []byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, 0, fmt.Errorf("new request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return c.Do(op, req)
}

// Do performs the request and returns status, body and the response's
// Retry-After in seconds (0 when absent or unparseable). op is the provider
// operation name reported to the metrics callback. A 429 or 503 answers with
// the cooldown the server wants; honouring it beats guessing a backoff.
func (c *Client) Do(op string, req *http.Request) (int, []byte, int, error) {
	req.Header.Set("User-Agent", c.userAgent)
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	var lastErr error
	var lastStatus int
	var lastBody []byte
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * RetryBackoff)
		}
		if c.limiter != nil {
			if err := c.limiter.Wait(req.Context()); err != nil {
				return 0, nil, 0, fmt.Errorf("%w: %w", ErrLimiterWait, err)
			}
		}
		start := time.Now()
		resp, err := c.base.Do(req)
		if err != nil {
			lastErr = err
			if c.metrics != nil {
				c.metrics(op, time.Since(start), err)
			}
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, c.maxBody+1))
		resp.Body.Close()
		if readErr == nil && int64(len(body)) > c.maxBody {
			readErr = fmt.Errorf("%w: limit %d bytes", ErrBodyTooLarge, c.maxBody)
		}
		if c.metrics != nil {
			c.metrics(op, time.Since(start), readErr)
		}
		if errors.Is(readErr, ErrBodyTooLarge) {
			return resp.StatusCode, nil, 0, readErr
		}
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= ServerErrThreshold {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			lastStatus, lastBody = resp.StatusCode, body
			continue
		}
		return resp.StatusCode, body, retryAfterSeconds(resp), nil
	}
	return lastStatus, lastBody, 0, fmt.Errorf("after %d attempts: %w", c.maxRetries, lastErr)
}

// retryAfterSeconds reads the Retry-After header. Only the delta-seconds form
// is honoured — the HTTP-date form is rare in practice and a misparsed date
// would produce a nonsense cooldown, so an unparseable value reports 0 and the
// caller falls back to its own default.
func retryAfterSeconds(resp *http.Response) int {
	v := resp.Header.Get("Retry-After")
	if v == "" {
		return 0
	}
	secs, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || secs < 0 {
		return 0
	}
	return secs
}
