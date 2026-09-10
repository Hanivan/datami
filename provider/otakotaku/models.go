package otakotaku

type otakotakuResponse struct {
	Data struct {
		ID      int    `json:"id_anime"`
		Title   string `json:"judul_anime"`
		MalID   *int   `json:"mal_id_anime"`
		AnidbID *int   `json:"anidb_id_anime"`
		AnnID   *int   `json:"ann_id_anime"`
		ApID    *int   `json:"ap_id_anime"`
	} `json:"data"`
}
