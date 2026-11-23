package anilist

import "strconv"

// desc is the description field selection. nil asHTML means HTML, so Media
// and Character descriptions share one format by default.
func (c *Client) desc() string {
	return "description(asHtml: " + strconv.FormatBool(c.asHTML == nil || *c.asHTML) + ")"
}

// animeSel is the media selection of GetByMalID, shared with the batch lookup.
func (c *Client) animeSel() string {
	return `id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format episodes duration chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } studios { edges { isMain node { id name isAnimationStudio } } } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } }`
}

// mangaSel is the media selection of GetMangaByMalID, shared with the batch lookup.
func (c *Client) mangaSel() string {
	return `id type title { romaji english native } ` + c.desc() + ` startDate { year month day } endDate { year month day } status format chapters volumes averageScore meanScore popularity favourites genres tags { name category rank isGeneralSpoiler isMediaSpoiler isAdult } coverImage { medium large } bannerImage relations { edges { relationType node { id idMal type format title { romaji } } } } synonyms idMal source countryOfOrigin isAdult externalLinks { id url site siteId type language notes isDisabled } rankings { id rank type format year season allTime context } stats { scoreDistribution { score amount } statusDistribution { status amount } }`
}

// charBody is the selection inside a characters connection, shared by the
// standalone and embedded queries.
func (c *Client) charBody() string {
	return `{ pageInfo { hasNextPage } edges { role node { id name { first middle last full native alternative } image { medium large } ` + c.desc() + ` } voiceActors { id name { first middle last full native alternative } image { medium large } languageV2 } } }`
}
