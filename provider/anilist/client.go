package anilist

import (
	"context"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/internal/httpx"
	"github.com/Hanivan/datami/types"
)

func (c *Client) Get(ctx context.Context, id string, opts ...types.MediaOpts) (*types.Media, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpGet, types.ErrInvalidInput)
	}
	query := `query ($id: Int) { Media(id: $id) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format episodes duration chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { edges { isMain node { id name isAnimationStudio } } } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } }`
	vars := map[string]any{"id": parseID(id)}
	return c.doGraphQL(ctx, query, vars, OpGet, opts)
}

func (c *Client) Search(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpSearch, types.ErrInvalidInput)
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	gql := `query ($search: String, $perPage: Int) { Page(perPage: $perPage) { Media(search: $search) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format episodes duration averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { edges { isMain node { id name isAnimationStudio } } } coverImage { medium large } bannerImage synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } } }`
	vars := map[string]any{"search": query, "perPage": limit}
	return c.doSearchGraphQL(ctx, gql, vars)
}

// GetByMalID fetches the ANIME whose MAL id is malID. The type filter is not
// optional: MAL anime and manga ids are separate namespaces, so an unfiltered
// idMal lookup can resolve to the manga that happens to share the number. Use
// GetMangaByMalID for the manga namespace.
func (c *Client) GetByMalID(ctx context.Context, malID string, opts ...types.MediaOpts) (*types.Media, error) {
	if malID == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpGetByMalID, types.ErrInvalidInput)
	}
	query := `query ($malId: Int, $type: MediaType) { Media(idMal: $malId, type: $type) { ` + c.animeSel() + ` } }`
	vars := map[string]any{"malId": parseID(malID), "type": "ANIME"}
	return c.doGraphQL(ctx, query, vars, OpGetByMalID, opts)
}

func (c *Client) GetManga(ctx context.Context, id string, opts ...types.MediaOpts) (*types.Media, error) {
	if id == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpGetManga, types.ErrInvalidInput)
	}
	query := `query ($id: Int) { Media(id: $id, type: MANGA) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } }`
	vars := map[string]any{"id": parseID(id)}
	return c.doGraphQL(ctx, query, vars, OpGetManga, opts)
}

func (c *Client) SearchManga(ctx context.Context, query string, opts types.SearchOpts) ([]types.Media, error) {
	if query == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpSearchManga, types.ErrInvalidInput)
	}
	limit := opts.Limit
	if limit == 0 {
		limit = httpx.DefaultSearchLimit
	}
	gql := `query ($search: String, $perPage: Int) { Page(perPage: $perPage) { Media(search: $search, type: MANGA) { id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } } } } }`
	vars := map[string]any{"search": query, "perPage": limit}
	return c.doSearchGraphQL(ctx, gql, vars)
}

func (c *Client) GetMangaByMalID(ctx context.Context, malID string, opts ...types.MediaOpts) (*types.Media, error) {
	if malID == "" {
		return nil, helpers.Wrap(types.SourceAniList, OpGetMangaByMalID, types.ErrInvalidInput)
	}
	query := `query ($malId: Int) { Media(idMal: $malId, type: MANGA) { ` + c.mangaSel() + ` } }`
	vars := map[string]any{"malId": parseID(malID)}
	return c.doGraphQL(ctx, query, vars, OpGetMangaByMalID, opts)
}
