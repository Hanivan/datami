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
	// DefaultMaxBody caps a response body read into memory.
	DefaultMaxBody = 32 << 20
)

// ErrBodyTooLarge marks a response body over Config.MaxBody. Not retried.
var ErrBodyTooLarge = errors.New("httpx: response body too large")

// ErrLimiterWait marks a failed rate limiter wait: no request was sent. The
// limiter or ctx error stays in the chain.
var ErrLimiterWait = errors.New("httpx: rate limiter wait")

type Config struct {
	Timeout   time.Duration
	Rate      float64
	Burst     int
	Metrics   func(op string, dur time.Duration, err error)
	Transport http.RoundTripper
	Headers   map[string]string
	// MaxRetries is total attempts per request; 0 uses the MaxRetries const.
	MaxRetries int
	// MaxBody caps the response body in bytes; 0 uses DefaultMaxBody.
	MaxBody int64
}

type Client struct {
	base       *http.Client
	limiter    *rate.Limiter
	metrics    func(op string, dur time.Duration, err error)
	headers    map[string]string
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
	var limiter *rate.Limiter
	if cfg.Rate > 0 {
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
		maxRetries: maxRetries,
		maxBody:    maxBody,
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

// Do performs the request and returns the status and body. Retry-After, when
// the server sends one, is discarded — use DoWithRetryAfter to see it.
func (c *Client) Do(req *http.Request) (int, []byte, error) {
	code, body, _, err := c.DoWithRetryAfter(req)
	return code, body, err
}

// DoWithRetryAfter is Do plus the response's Retry-After value in seconds (0
// when absent or unparseable). A 429 or 503 answers with the cooldown the
// server actually wants; honouring it beats guessing a backoff, since guessing
// short keeps the rate limit alive.
func (c *Client) DoWithRetryAfter(req *http.Request) (int, []byte, int, error) {
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
				c.metrics(req.Method, time.Since(start), err)
			}
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, c.maxBody+1))
		resp.Body.Close()
		if readErr == nil && int64(len(body)) > c.maxBody {
			readErr = fmt.Errorf("%w: limit %d bytes", ErrBodyTooLarge, c.maxBody)
		}
		if c.metrics != nil {
			c.metrics(req.Method, time.Since(start), readErr)
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
