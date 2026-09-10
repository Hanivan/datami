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
	ID    string
	Names []string
	Image string
	// ImageLarge is the large variant when the source supplies both sizes;
	// Image holds the medium/default one.
	ImageLarge      string
	Biography       string
	FirstName       string
	LastName        string
	NativeName      string
	AlternativeName []string
	// NumFavorites is the source's favourite count, nil when not supplied.
	// A pointer because 0 is a real count and must not read as absent.
	NumFavorites *int
	Roles        []CharacterRole
	// VoiceActors are the character's VAs for the media it was fetched with
	// (AniList only; MAL's character roster carries none).
	VoiceActors []VoiceActor
}

// VoiceActor is one voice credit on a Character, with the language it was
// performed in. Language separates the JP cast from dubs — the consumer keys
// its edge rows on (person, language), so it cannot be dropped.
type VoiceActor struct {
	ID         string
	Name       string
	NativeName string
	Language   string
	Image      string
	ImageLarge string
}

// CharacterRole links a character to a media with a role (main/supporting).
type CharacterRole struct {
	MediaID string
	Role    CharacterRoleType
}

// Person is a canonical person entity (voice actor, staff, author).
type Person struct {
	ID        string
	Name      string
	Image     string
	Birthday  string
	Biography string
	// ImageLarge is the large variant when the source supplies both sizes.
	ImageLarge string
	// NativeName is the name in its original script, when supplied.
	NativeName string
	// Role is the staff credit this person holds on the media they were
	// fetched with ("Story", "Art", ...); empty when fetched standalone.
	Role string
	// PrimaryOccupations is AniList's occupation list ("Mangaka", ...).
	PrimaryOccupations []string
	// NumFavorites is the source's favourite count, nil when not supplied.
	NumFavorites *int
}
