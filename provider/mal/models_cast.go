package mal

type character struct {
	ID              int      `json:"id"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	AlternativeName string   `json:"alternative_name,omitempty"`
	MainPicture     *picture `json:"main_picture,omitempty"`
	Biography       string   `json:"biography,omitempty"`
	// NumFavorites is a pointer: 0 is a real favourite count and must not be
	// indistinguishable from a field MAL did not return.
	NumFavorites *int `json:"num_favorites,omitempty"`
}

type animeCharacter struct {
	Node character `json:"node"`
	Role string    `json:"role"`
}

type person struct {
	ID           int      `json:"id"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	MainPicture  *picture `json:"main_picture,omitempty"`
	Birthday     string   `json:"birthday,omitempty"`
	NumFavorites *int     `json:"num_favorites,omitempty"`
	More         string   `json:"more,omitempty"`
}
