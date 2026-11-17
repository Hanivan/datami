package mal

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

func (c *Client) GetAnimeCharacters(ctx context.Context, animeID string) ([]types.Character, error) {
	if animeID == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/anime/%s/characters?limit=100&fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography", c.baseURL, url.PathEscape(animeID))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetAnimeCharacters, Err: err}
	}
	return handleCharactersResponse(code, body)
}

func (c *Client) GetCharacter(ctx context.Context, charID string) (*types.Character, error) {
	if charID == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/characters/%s?fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography,animeography", c.baseURL, url.PathEscape(charID))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetCharacter, Err: err}
	}
	return handleCharacterResponse(code, body)
}

func (c *Client) GetPerson(ctx context.Context, personID string) (*types.Person, error) {
	if personID == "" {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: types.ErrInvalidInput}
	}
	u := fmt.Sprintf("%s/people/%s?fields=first_name,last_name,main_picture,birthday,num_favorites,more", c.baseURL, url.PathEscape(personID))
	code, body, err := c.client.Get(ctx, u)
	if err != nil {
		return nil, &types.Error{Source: types.SourceMAL, Op: OpGetPerson, Err: err}
	}
	return handlePersonResponse(code, body)
}

func handleCharactersResponse(code int, body []byte) ([]types.Character, error) {
	resp, err := helpers.DecodeJSON[charactersResponse](types.SourceMAL, OpGetAnimeCharacters, code, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Character, 0, len(resp.Data))
	for _, ac := range resp.Data {
		results = append(results, *ac.toCharacter())
	}
	return results, nil
}

func handleCharacterResponse(code int, body []byte) (*types.Character, error) {
	c, err := helpers.DecodeJSON[character](types.SourceMAL, OpGetCharacter, code, body)
	if err != nil {
		return nil, err
	}
	return c.toCharacter(), nil
}

func handlePersonResponse(code int, body []byte) (*types.Person, error) {
	p, err := helpers.DecodeJSON[person](types.SourceMAL, OpGetPerson, code, body)
	if err != nil {
		return nil, err
	}
	return p.toPerson(), nil
}

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
