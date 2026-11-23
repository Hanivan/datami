package anilist

import (
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

func (e characterEdge) toCharacter() *types.Character {
	c := &types.Character{
		ID:              strconv.Itoa(e.Node.ID),
		Biography:       helpers.DerefString(e.Node.Description),
		FirstName:       e.Node.Name.First,
		LastName:        e.Node.Name.Last,
		NativeName:      e.Node.Name.Native,
		AlternativeName: e.Node.Name.Alternative,
		Image:           e.Node.Image.Medium,
		ImageLarge:      e.Node.Image.Large,
		Roles: []types.CharacterRole{
			{MediaID: "", Role: types.CharacterRoleType(strings.ToLower(e.Role))},
		},
	}
	name := e.Node.Name.Full
	if name == "" {
		name = strings.TrimSpace(e.Node.Name.First + " " + e.Node.Name.Last)
	}
	c.Names = append(c.Names, name)
	if c.Image == "" {
		c.Image = e.Node.Image.Large
	}
	for _, va := range e.VoiceActors {
		vaName := va.Name.Full
		if vaName == "" {
			vaName = va.Name.Native
		}
		c.VoiceActors = append(c.VoiceActors, types.VoiceActor{
			ID:         strconv.Itoa(va.ID),
			Name:       vaName,
			NativeName: va.Name.Native,
			Language:   va.LanguageV2,
			Image:      va.Image.Medium,
			ImageLarge: va.Image.Large,
		})
	}
	return c
}

func (e staffEdge) toPerson() *types.Person {
	name := e.Node.Name.Full
	if name == "" {
		name = e.Node.Name.Native
	}
	img := e.Node.Image.Large
	if img == "" {
		img = e.Node.Image.Medium
	}
	return &types.Person{
		ID:                 strconv.Itoa(e.Node.ID),
		Name:               name,
		NativeName:         e.Node.Name.Native,
		Role:               e.Role,
		Image:              img,
		ImageLarge:         e.Node.Image.Large,
		PrimaryOccupations: e.Node.PrimaryOccupations,
	}
}
