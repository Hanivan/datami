package types

// UserInfo is the authenticated AniList viewer.
type UserInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MediaListEntry is one entry in a user's AniList list.
type MediaListEntry struct {
	ID       int    `json:"id"`
	MediaID  int    `json:"mediaId"`
	UserID   int    `json:"userId"`
	Status   string `json:"status"`
	Score    int    `json:"score"`
	Progress int    `json:"progress"`
}

// MediaListCollection is a user's full AniList list grouped by status.
type MediaListCollection struct {
	Lists []MediaList `json:"lists"`
}

type MediaList struct {
	Name    string           `json:"name"`
	Entries []MediaListEntry `json:"entries"`
}
