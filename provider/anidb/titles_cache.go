package anidb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (c *Client) downloadTitles(ctx context.Context) ([]byte, error) {
	code, body, _, err := c.client.Get(ctx, OpSearch, c.titlesURL)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("status %d", code)
	}
	return body, nil
}

func readTitlesFile(path string) ([]titleRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseTitles(f)
}

// writeTitlesFile writes atomically (temp + rename) and stamps mtime = now.
func writeTitlesFile(path string, data []byte, now time.Time) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, titlesFile+".tmp*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chtimes(tmp, now, now)
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
	}
	return err
}
