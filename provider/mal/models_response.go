package mal

type searchResponse struct {
	Data   []searchNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type searchNode struct {
	Node anime `json:"node"`
}

type mangaSearchResponse struct {
	Data   []mangaSearchNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type mangaSearchNode struct {
	Node manga `json:"node"`
}

type seasonalResponse struct {
	Data   []searchNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type rankingResponse struct {
	Data   []rankingNode `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

type rankingNode struct {
	Node    anime `json:"node"`
	Ranking struct {
		Rank int `json:"rank"`
	} `json:"ranking"`
}

type charactersResponse struct {
	Data   []animeCharacter `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}
