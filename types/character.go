package types

// Character is a canonical character entity (anime/manga). Source-independent
// so enrich can map MAL/AniList character payloads onto one shape.
type Character struct {
	ID        string
	Names     []string
	Image     string
	Biography string
	Roles     []CharacterRole
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
}
