package anilist

// charConn is a paginated characters connection.
type charConn struct {
	PageInfo struct {
		HasNextPage bool `json:"hasNextPage"`
	} `json:"pageInfo"`
	Edges []characterEdge `json:"edges"`
}

// staffConn is a paginated staff connection.
type staffConn struct {
	PageInfo struct {
		HasNextPage bool `json:"hasNextPage"`
	} `json:"pageInfo"`
	Edges []staffEdge `json:"edges"`
}

type characterEdge struct {
	Role        string        `json:"role"`
	Node        characterNode `json:"node"`
	VoiceActors []voiceActor  `json:"voiceActors"`
}

type characterNode struct {
	ID          int            `json:"id"`
	Name        characterName  `json:"name"`
	Image       characterImage `json:"image"`
	Description *string        `json:"description,omitempty"`
}

type voiceActor struct {
	ID         int            `json:"id"`
	Name       characterName  `json:"name"`
	Image      characterImage `json:"image"`
	LanguageV2 string         `json:"languageV2"`
}

type characterName struct {
	First       string   `json:"first"`
	Middle      string   `json:"middle"`
	Last        string   `json:"last"`
	Full        string   `json:"full"`
	Native      string   `json:"native"`
	Alternative []string `json:"alternative"`
}

type characterImage struct {
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type staffEdge struct {
	Role string    `json:"role"`
	Node staffNode `json:"node"`
}

type staffNode struct {
	ID                 int            `json:"id"`
	Name               characterName  `json:"name"`
	Image              characterImage `json:"image"`
	PrimaryOccupations []string       `json:"primaryOccupations"`
}
