package datami

import (
	"fmt"
	"time"

	"github.com/Hanivan/datami/provider/anidb"
	"github.com/Hanivan/datami/provider/anilist"
	"github.com/Hanivan/datami/provider/kitsu"
	"github.com/Hanivan/datami/provider/mal"
)

type Client struct {
	MAL     *mal.Client
	AniList *anilist.Client
	AniDB   *anidb.Client
	Kitsu   *kitsu.Client
}

func New(cfg Config) (*Client, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	metrics := cfg.Metrics

	malClient, err := mal.New(mal.Config{
		ClientID: cfg.MALClientID,
		Timeout:  timeout,
		Metrics:  metrics,
	})
	if err != nil {
		return nil, fmt.Errorf("mal: %w", err)
	}

	anilistClient, err := anilist.New(anilist.Config{
		Token:             cfg.AniListToken,
		Timeout:           timeout,
		Metrics:           metrics,
		DescriptionAsHTML: cfg.AniListDescriptionAsHTML,
	})
	if err != nil {
		return nil, fmt.Errorf("anilist: %w", err)
	}

	anidbClient := anidb.Unconfigured()
	if cfg.AniDBClientName != "" {
		anidbClient, err = anidb.New(anidb.Config{
			ClientName:    cfg.AniDBClientName,
			ClientVersion: cfg.AniDBClientVersion,
			CacheDir:      cfg.AniDBCacheDir,
			Timeout:       timeout,
			Metrics:       metrics,
		})
		if err != nil {
			return nil, err
		}
	}

	kitsuClient, err := kitsu.New(kitsu.Config{
		Timeout: timeout,
		Metrics: metrics,
	})
	if err != nil {
		return nil, fmt.Errorf("kitsu: %w", err)
	}

	return &Client{
		MAL:     malClient,
		AniList: anilistClient,
		AniDB:   anidbClient,
		Kitsu:   kitsuClient,
	}, nil
}
