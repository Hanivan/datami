package anidb

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

type Config struct {
	// ClientName is a registered AniDB HTTP client name. Required.
	ClientName string
	// ClientVersion is the registered client version. Required, > 0.
	ClientVersion int
	// CacheDir holds the titles dump; default os.UserCacheDir()/datami/anidb.
	CacheDir string
	Timeout  time.Duration
	Metrics  func(source string, op string, dur time.Duration, err error)
}

type Client struct {
	apiURL        string
	titlesURL     string
	clientName    string
	clientVersion int
	cacheDir      string
	client        *httpx.Client
	now           func() time.Time
	metrics       func(source, op string, dur time.Duration, err error)
	titles        titlesState
	// unconfigured makes every method fail with ErrInvalidInput, no network.
	unconfigured bool
}

// New returns ErrInvalidInput when ClientName is empty or ClientVersion <= 0.
func New(cfg Config) (*Client, error) {
	if cfg.ClientName == "" || cfg.ClientVersion <= 0 {
		return nil, wrap(OpNew, types.ErrInvalidInput)
	}
	if cfg.CacheDir == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return nil, wrap(OpNew, err)
		}
		cfg.CacheDir = filepath.Join(dir, "datami", "anidb")
	}
	return newClient(defaultAPIURL, defaultTitlesURL, cfg), nil
}

// Unconfigured returns a client whose every method returns ErrInvalidInput
// without touching the network. Used when no AniDB client is registered.
func Unconfigured() *Client {
	return &Client{unconfigured: true, now: time.Now, titles: newTitlesState()}
}

// NewWithTestServer points the client at test URLs. Empty ClientName is
// replaced by "test"/1 so tests need not set it.
func NewWithTestServer(apiURL, titlesURL string, cfg Config) *Client {
	if cfg.ClientName == "" {
		cfg.ClientName, cfg.ClientVersion = "test", 1
	}
	return newClient(apiURL, titlesURL, cfg)
}

func newClient(apiURL, titlesURL string, cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = httpx.DefaultTimeout
	}
	return &Client{
		apiURL:        apiURL,
		titlesURL:     titlesURL,
		clientName:    cfg.ClientName,
		clientVersion: cfg.ClientVersion,
		cacheDir:      cfg.CacheDir,
		now:           time.Now,
		metrics:       cfg.Metrics,
		titles:        newTitlesState(),
		client: httpx.New(httpx.Config{
			Timeout:    timeout,
			Rate:       0.5,
			Burst:      1,
			MaxRetries: 1,
			// The gzipped titles dump is ~10 MiB; leave headroom over the default.
			MaxBody: 64 << 20,
			Metrics: httpx.SourceMetrics(types.SourceAniDB, cfg.Metrics),
		}),
	}
}
