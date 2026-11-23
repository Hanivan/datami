package mal

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

func (c *Client) GetAnimeCharacters(ctx context.Context, animeID string) ([]types.Character, error) {
	if animeID == "" {
		return nil, helpers.Wrap(types.SourceMAL, OpGetAnimeCharacters, types.ErrInvalidInput)
	}
	u := fmt.Sprintf("%s/anime/%s/characters?limit=100&fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography", c.baseURL, url.PathEscape(animeID))
	code, body, retryAfter, err := c.client.Get(ctx, OpGetAnimeCharacters, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGetAnimeCharacters, err)
	}
	return handleCharactersResponse(code, body, retryAfter)
}

func (c *Client) GetCharacter(ctx context.Context, charID string) (*types.Character, error) {
	if charID == "" {
		return nil, helpers.Wrap(types.SourceMAL, OpGetCharacter, types.ErrInvalidInput)
	}
	u := fmt.Sprintf("%s/characters/%s?fields=first_name,last_name,alternative_name,main_picture,num_favorites,biography,animeography", c.baseURL, url.PathEscape(charID))
	code, body, retryAfter, err := c.client.Get(ctx, OpGetCharacter, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGetCharacter, err)
	}
	return handleCharacterResponse(code, body, retryAfter)
}

func (c *Client) GetPerson(ctx context.Context, personID string) (*types.Person, error) {
	if personID == "" {
		return nil, helpers.Wrap(types.SourceMAL, OpGetPerson, types.ErrInvalidInput)
	}
	u := fmt.Sprintf("%s/people/%s?fields=first_name,last_name,main_picture,birthday,num_favorites,more", c.baseURL, url.PathEscape(personID))
	code, body, retryAfter, err := c.client.Get(ctx, OpGetPerson, u)
	if err != nil {
		return nil, helpers.Wrap(types.SourceMAL, OpGetPerson, err)
	}
	return handlePersonResponse(code, body, retryAfter)
}
