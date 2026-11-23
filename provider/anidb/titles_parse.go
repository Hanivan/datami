package anidb

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/Hanivan/datami/types"
)

// parseTitles streams a gzip (sniffed by magic bytes) or plain XML dump.
func parseTitles(r io.Reader) ([]titleRecord, error) {
	return parseTitlesN(r, maxTitlesDump)
}

func (c *countReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// parseTitlesN is parseTitles with a cap on decompressed bytes.
func parseTitlesN(r io.Reader, max int64) ([]titleRecord, error) {
	br := bufio.NewReader(r)
	var src io.Reader = br
	if magic, _ := br.Peek(2); len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		zr, err := gzip.NewReader(br)
		if err != nil {
			return nil, fmt.Errorf("gzip: %w", err)
		}
		defer zr.Close()
		src = zr
	}
	cr := &countReader{r: io.LimitReader(src, max+1)}
	var d titlesDump
	err := xml.NewDecoder(cr).Decode(&d)
	if err == nil {
		io.Copy(io.Discard, cr) // count trailing bytes too
	}
	if cr.n > max {
		return nil, fmt.Errorf("parse titles: %w", errTooLarge)
	}
	if err != nil {
		return nil, fmt.Errorf("parse titles: %w", err)
	}
	if len(d.Anime) == 0 {
		return nil, fmt.Errorf("%w: empty titles dump", types.ErrUnavailable)
	}
	idx := make([]titleRecord, 0, len(d.Anime))
	for _, a := range d.Anime {
		r := titleRecord{aid: a.AID, raw: titles{Title: a.Titles}}
		for _, t := range a.Titles {
			r.norms = append(r.norms, normalizeTitle(t.Text))
		}
		idx = append(idx, r)
	}
	sort.Slice(idx, func(i, j int) bool { return idx[i].aid < idx[j].aid })
	return idx, nil
}

// normalizeTitle lowercases, keeps letters/digits, collapses the rest to
// single spaces.
func normalizeTitle(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(r)
		} else {
			space = true
		}
	}
	return b.String()
}
