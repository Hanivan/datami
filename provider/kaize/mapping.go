package kaize

import (
	"bytes"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
	"golang.org/x/net/html"
)

func parseKaizeList(body []byte) ([]kaizeItem, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	var items []kaizeItem
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && helpers.HasClass(n, "anime-list-element") {
			item := extractAnimeElement(n)
			if item.Title != "" {
				items = append(items, item)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return items, nil
}

func extractAnimeElement(n *html.Node) kaizeItem {
	var item kaizeItem
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "name") {
					item.Title = helpers.TextContent(n)
					if item.Slug == "" {
						for _, ha := range n.Attr {
							if ha.Key == "href" {
								parts := strings.Split(ha.Val, "/")
								item.Slug = parts[len(parts)-1]
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return item
}

func (i kaizeItem) toMedia() *types.Media {
	return &types.Media{
		Titles: []types.Title{{Text: i.Title}},
		Raw: map[string]any{
			"slug":  i.Slug,
			"kaize": i.Slug,
		},
	}
}
