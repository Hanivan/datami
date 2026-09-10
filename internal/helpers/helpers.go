package helpers

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Hanivan/datami/types"
	"golang.org/x/net/html"
)

// ParseDate parses a "YYYY-MM-DD" string into a types.Date.
func ParseDate(s string) *types.Date {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "-")
	if len(parts) == 0 {
		return nil
	}
	d := &types.Date{}
	if y, err := strconv.Atoi(parts[0]); err == nil {
		d.Year = &y
	}
	if len(parts) >= 2 {
		if m, err := strconv.Atoi(parts[1]); err == nil {
			d.Month = &m
		}
	}
	if len(parts) >= 3 {
		if day, err := strconv.Atoi(parts[2]); err == nil {
			d.Day = &day
		}
	}
	return d
}

// RawPayload marshals v into JSON and back into map[string]any.
func RawPayload(v any) map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// NormalizeRelation normalizes a relation string (e.g. "Side Story" -> "side_story").
func NormalizeRelation(rel string) string {
	return strings.ToLower(strings.ReplaceAll(rel, " ", "_"))
}

// DerefString returns the value of s or "" if nil.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefInt returns the value of i or 0 if nil.
func DerefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// TextContent extracts all text content from an HTML node.
func TextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(TextContent(c))
	}
	return strings.TrimSpace(sb.String())
}

// HasClass checks if an HTML node has the given class.
func HasClass(n *html.Node, class string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" && strings.Contains(a.Val, class) {
			return true
		}
	}
	return false
}

// TdNodes collects all <td> child nodes from an HTML node.
func TdNodes(n *html.Node) []*html.Node {
	var cols []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "td" {
			cols = append(cols, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return cols
}
