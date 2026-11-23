package anidb

import (
	"strconv"
	"strings"

	"github.com/Hanivan/datami/types"
)

// toCharacters maps the characters block, keeping only charactertype id 1.
// ponytail: secondary/appears/cameo all map to RoleSupporting (enum has two values).
func (a anime) toCharacters(aid string) []types.Character {
	var out []types.Character
	for _, ch := range a.Characters {
		if ch.CharacterType.ID != characterTypeCharacter {
			continue
		}
		role := types.RoleSupporting
		if ch.Type == "main character in" {
			role = types.RoleMain
		}
		c := types.Character{
			ID:        strconv.Itoa(ch.ID),
			Names:     []string{ch.Name},
			Gender:    ch.Gender,
			Biography: ch.Description,
			Image:     cdnImage(ch.Picture),
			Roles:     []types.CharacterRole{{MediaID: aid, Role: role}},
		}
		if s := ch.Seiyuu; s != nil {
			c.VoiceActors = []types.VoiceActor{{
				ID: strconv.Itoa(s.ID), Name: strings.TrimSpace(s.Name), Language: "ja", Image: cdnImage(s.Picture),
			}}
		}
		out = append(out, c)
	}
	return out
}
