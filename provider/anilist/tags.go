package anilist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// GetTagsAndLinks returns a media's tags and external links by AniList id.
func (c *Client) GetTagsAndLinks(ctx context.Context, id string) (*TagsLinks, error) {
	return c.tagsLinksBy(ctx, "id", id, "")
}

// GetTagsAndLinksByMalID returns a media's tags and external links addressed by
// its MAL id. See GetCharactersByMalID on why the id namespace matters.
func (c *Client) GetTagsAndLinksByMalID(ctx context.Context, malID, mediaType string) (*TagsLinks, error) {
	return c.tagsLinksBy(ctx, "idMal", malID, mediaType)
}

func (c *Client) tagsLinksBy(ctx context.Context, idField, id, mediaType string) (*TagsLinks, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpGetTagsLinks, types.ErrInvalidInput)
	}
	sel := `tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } externalLinks { id url site siteId type language notes isDisabled }`
	vars := map[string]any{idField: parseID(id)}
	args := "$" + idField + ": Int"
	mediaArgs := idField + ": $" + idField
	if mediaType != "" {
		args += ", $type: MediaType"
		mediaArgs += ", type: $type"
		vars["type"] = mediaType
	}
	query := "query (" + args + ") { Media(" + mediaArgs + ") { " + sel + " } }"
	body, err := c.post(ctx, query, vars, OpGetTagsLinks)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data struct {
			Media *struct {
				Tags          []tag          `json:"tags"`
				ExternalLinks []externalLink `json:"externalLinks"`
			} `json:"Media"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, helpers.Wrap(types.SourceAniList, OpGetTagsLinks, fmt.Errorf("decode: %w", err))
	}
	if result.Data.Media == nil {
		return nil, helpers.Wrap(types.SourceAniList, OpGetTagsLinks, ErrMediaNotFound)
	}
	tl := &TagsLinks{}
	for _, t := range result.Data.Media.Tags {
		tl.Tags = append(tl.Tags, Tag{
			Name: t.Name, Category: t.Category, Rank: t.Rank,
			IsGeneralSpoiler: t.IsGeneralSpoiler, IsMediaSpoiler: t.IsMediaSpoiler, IsAdult: t.IsAdult,
		})
	}
	for _, l := range result.Data.Media.ExternalLinks {
		tl.ExternalLinks = append(tl.ExternalLinks, ExternalLink{
			ID: l.ID, URL: l.URL, Site: l.Site, SiteID: l.SiteID, Type: l.Type,
			Language: l.Language, Notes: l.Notes, IsDisabled: l.IsDisabled,
		})
	}
	return tl, nil
}
