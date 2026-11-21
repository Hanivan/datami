package types

type SearchOpts struct {
	Limit     int
	MediaType string
	Year      int
	Status    string
	Page      int
}

// MediaOpts asks a media lookup to embed cast/staff pages in the same query.
// Zero values embed nothing. Pages are 25 edges each (AniList's cap).
// MaxCharacterPages and MaxStaffPages cap the total pages per connection,
// embedded plus follow-up; zero means 10.
type MediaOpts struct {
	CharacterPages    int
	StaffPages        int
	MaxCharacterPages int
	MaxStaffPages     int
}

// PageOpts bounds a paged character/staff walk. Zero StartPage means 1, zero
// MaxPages means the default cap of 10.
type PageOpts struct {
	StartPage int
	MaxPages  int
}
