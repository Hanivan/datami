package mal

import (
	"strconv"
	"strings"

	"github.com/Hanivan/datami/types"
)

func (a animeCharacter) toCharacter() *types.Character {
	c := a.Node.toCharacter()
	// The roster's role is MAL's VERBATIM string ("Main"/"Supporting"), kept
	// lowercased here; consumers that need a different casing normalize it.
	c.Roles = []types.CharacterRole{
		{MediaID: "", Role: types.CharacterRoleType(strings.ToLower(a.Role))},
	}
	return c
}

func (c character) toCharacter() *types.Character {
	out := &types.Character{
		ID:           strconv.Itoa(c.ID),
		Biography:    c.Biography,
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		NumFavorites: c.NumFavorites,
	}
	if c.AlternativeName != "" {
		out.AlternativeName = []string{c.AlternativeName}
	}
	if c.FirstName != "" || c.LastName != "" {
		out.Names = append(out.Names, strings.TrimSpace(c.FirstName+" "+c.LastName))
	}
	if c.MainPicture != nil {
		out.Image = c.MainPicture.Medium
		out.ImageLarge = c.MainPicture.Large
	}
	return out
}

func (p person) toPerson() *types.Person {
	out := &types.Person{
		ID:           strconv.Itoa(p.ID),
		Name:         strings.TrimSpace(p.FirstName + " " + p.LastName),
		Birthday:     p.Birthday,
		Biography:    p.More,
		NumFavorites: p.NumFavorites,
	}
	if p.MainPicture != nil {
		out.Image = p.MainPicture.Medium
		out.ImageLarge = p.MainPicture.Large
	}
	return out
}
