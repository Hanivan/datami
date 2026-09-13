package types

// Character is a canonical character entity (anime/manga). Source-independent
// so enrich can map MAL/AniList character payloads onto one shape.
//
// The name parts are kept SEPARATE rather than pre-joined: consumers build
// search keys from the individual parts and persist first/last/native as
// distinct columns, which a single joined string cannot be split back into
// (a native name is not "first last", and an alternative list is not one
// value). Names stays as the display list.
type Character struct {
	ID    string   `json:"id"`
	Names []string `json:"names,omitempty"`
	Image string   `json:"image,omitempty"`
	// ImageLarge is the large variant when the source supplies both sizes;
	// Image holds the medium/default one.
	ImageLarge      string   `json:"image_large,omitempty"`
	Biography       string   `json:"biography,omitempty"`
	FirstName       string   `json:"first_name,omitempty"`
	LastName        string   `json:"last_name,omitempty"`
	NativeName      string   `json:"native_name,omitempty"`
	AlternativeName []string `json:"alternative_name,omitempty"`
	// NumFavorites is the source's favourite count, nil when not supplied.
	// A pointer because 0 is a real count and must not read as absent.
	NumFavorites *int            `json:"num_favorites,omitempty"`
	Roles        []CharacterRole `json:"roles,omitempty"`
	// VoiceActors are the character's VAs for the media it was fetched with
	// (AniList only; MAL's character roster carries none).
	VoiceActors []VoiceActor `json:"voice_actors,omitempty"`
}

// VoiceActor is one voice credit on a Character, with the language it was
// performed in. Language separates the JP cast from dubs — the consumer keys
// its edge rows on (person, language), so it cannot be dropped.
type VoiceActor struct {
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	NativeName string `json:"native_name,omitempty"`
	Language   string `json:"language,omitempty"`
	Image      string `json:"image,omitempty"`
	ImageLarge string `json:"image_large,omitempty"`
}

// CharacterRole links a character to a media with a role (main/supporting).
type CharacterRole struct {
	MediaID string            `json:"media_id"`
	Role    CharacterRoleType `json:"role"`
}

// Person is a canonical person entity (voice actor, staff, author).
type Person struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	Image     string `json:"image,omitempty"`
	Birthday  string `json:"birthday,omitempty"`
	Biography string `json:"biography,omitempty"`
	// ImageLarge is the large variant when the source supplies both sizes.
	ImageLarge string `json:"image_large,omitempty"`
	// NativeName is the name in its original script, when supplied.
	NativeName string `json:"native_name,omitempty"`
	// Role is the staff credit this person holds on the media they were
	// fetched with ("Story", "Art", ...); empty when fetched standalone.
	Role string `json:"role,omitempty"`
	// PrimaryOccupations is AniList's occupation list ("Mangaka", ...).
	PrimaryOccupations []string `json:"primary_occupations,omitempty"`
	// NumFavorites is the source's favourite count, nil when not supplied.
	NumFavorites *int `json:"num_favorites,omitempty"`
}
