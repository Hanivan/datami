package hikka

type hikkaResponse struct {
	Pagination struct {
		Total int `json:"total"`
		Pages int `json:"pages"`
	} `json:"pagination"`
	List []hikkaItem `json:"list"`
}

type hikkaItem struct {
	Slug    string `json:"slug"`
	MalID   int    `json:"mal_id"`
	TitleJA string `json:"title_ja"`
	TitleEN string `json:"title_en"`
	TitleUA string `json:"title_ua"`
}
