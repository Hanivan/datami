package nautiljon

import (
	"bytes"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
	"golang.org/x/net/html"
)

func normalizeStatus(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func parseNautiljonList(body []byte) ([]nautiljonItem, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var items []nautiljonItem
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "search") {
					items = extractTableRows(n)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return items, nil
}

func extractTableRows(table *html.Node) []nautiljonItem {
	var items []nautiljonItem
	var rows []*html.Node
	var findRows func(*html.Node)
	findRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			rows = append(rows, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findRows(c)
		}
	}
	findRows(table)
	for _, row := range rows {
		cols := helpers.TdNodes(row)
		if len(cols) >= 4 {
			title := helpers.TextContent(cols[1])
			format := helpers.TextContent(cols[2])
			status := helpers.TextContent(cols[3])
			items = append(items, nautiljonItem{
				Title:  title,
				Format: format,
				Status: status,
			})
		}
	}
	return items
}

func (i nautiljonItem) toMedia() *types.Media {
	m := &types.Media{
		Titles: []types.Title{{Text: i.Title}},
		Raw: map[string]any{
			"slug":   i.Slug,
			"format": i.Format,
			"status": i.Status,
		},
	}
	if s, ok := nautiljonStatusMap[normalizeStatus(i.Status)]; ok {
		m.Status = types.MediaStatus(s)
	}
	return m
}
