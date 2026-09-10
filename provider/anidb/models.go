package anidb

import "encoding/xml"

type anime struct {
	XMLName     xml.Name  `xml:"anime"`
	ID          int       `xml:"id,attr"`
	Titles      titles    `xml:"titles"`
	Description string    `xml:"description"`
	Episodes    int       `xml:"episodes"`
	StartDate   string    `xml:"startdate"`
	EndDate     string    `xml:"enddate"`
	Type        string    `xml:"type"`
	Picture     string    `xml:"picture"`
	Tags        []tag     `xml:"tags>tag"`
	Related     []related `xml:"related>anime"`
}

type titles struct {
	Title []titleEntry `xml:"title"`
}

type titleEntry struct {
	Text string `xml:",chardata"`
	Lang string `xml:"xml:lang,attr"`
	Type string `xml:"type,attr"`
}

type tag struct {
	Name string `xml:"name,attr"`
	ID   int    `xml:"id,attr"`
}

type related struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:",chardata"`
}

type searchResponse struct {
	XMLName xml.Name `xml:"anime"`
	Anime   []anime  `xml:"anime"`
}
