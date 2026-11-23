package anidb

import (
	"encoding/xml"
	"io"
	"time"
)

// XML structures adapted from go-anidb (MIT, see NOTICE).
type anime struct {
	XMLName      xml.Name    `xml:"anime"`
	ID           int         `xml:"id,attr"`
	Restricted   bool        `xml:"restricted,attr"`
	Type         string      `xml:"type"`
	EpisodeCount int         `xml:"episodecount"`
	StartDate    string      `xml:"startdate"`
	EndDate      string      `xml:"enddate"`
	Titles       titles      `xml:"titles"`
	Related      []related   `xml:"relatedanime>anime"`
	URL          string      `xml:"url"`
	Creators     []creator   `xml:"creators>name"`
	Description  string      `xml:"description"`
	Ratings      ratings     `xml:"ratings"`
	Picture      string      `xml:"picture"`
	Resources    []resource  `xml:"resources>resource"`
	Tags         []tag       `xml:"tags>tag"`
	Characters   []character `xml:"characters>character"`
}

type titles struct {
	Title []titleEntry `xml:"title"`
}

type titleEntry struct {
	Text string `xml:",chardata"`
	Lang string `xml:"lang,attr"`
	Type string `xml:"type,attr"`
}

type tag struct {
	ID       int    `xml:"id,attr"`
	ParentID int    `xml:"parentid,attr"`
	Weight   int    `xml:"weight,attr"`
	Spoiler  bool   `xml:"localspoiler,attr"`
	Name     string `xml:"name"`
}

type related struct {
	ID   int    `xml:"id,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:",chardata"`
}

type creator struct {
	ID   int    `xml:"id,attr"`
	Type string `xml:"type,attr"`
	Name string `xml:",chardata"`
}

type rating struct {
	Count int     `xml:"count,attr"`
	Value float64 `xml:",chardata"`
}

type ratings struct {
	Permanent rating `xml:"permanent"`
	Temporary rating `xml:"temporary"`
	Review    rating `xml:"review"`
}

type resource struct {
	Type     int              `xml:"type,attr"`
	Entities []externalEntity `xml:"externalentity"`
}

type externalEntity struct {
	Identifiers []string `xml:"identifier"`
	URLs        []string `xml:"url"`
}

type character struct {
	ID            int           `xml:"id,attr"`
	Type          string        `xml:"type,attr"` // "main character in", "secondary cast in", ...
	Name          string        `xml:"name"`
	Gender        string        `xml:"gender"`
	CharacterType characterType `xml:"charactertype"`
	Description   string        `xml:"description"`
	Picture       string        `xml:"picture"`
	Seiyuu        *seiyuu       `xml:"seiyuu"`
}

type characterType struct {
	ID   int    `xml:"id,attr"` // 1 = Character; key off this, not the label
	Name string `xml:",chardata"`
}

type seiyuu struct {
	ID      int    `xml:"id,attr"`
	Picture string `xml:"picture,attr"`
	Name    string `xml:",chardata"`
}

// titlesDump is the anime-titles.xml layout (format per go-anidb, see NOTICE).
type titlesDump struct {
	Anime []struct {
		AID    int          `xml:"aid,attr"`
		Titles []titleEntry `xml:"title"`
	} `xml:"anime"`
}

type titleRecord struct {
	aid   int
	raw   titles
	norms []string
}

// titlesState is the in-memory index plus refresh bookkeeping. sem is a
// 1-slot lock guarding the rest; it is held across the download so concurrent
// Searches share one request, and waiters can give up via ctx.
type titlesState struct {
	sem         chan struct{}
	index       []titleRecord // sorted by aid
	loadedAt    time.Time
	lastAttempt time.Time
	lastErr     error
}

type countReader struct {
	r io.Reader
	n int64
}
