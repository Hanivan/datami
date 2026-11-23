package mal

import (
	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

func handleResponse(code int, body []byte, op string, retryAfter int) (*types.Media, error) {
	a, err := helpers.DecodeJSON[anime](types.SourceMAL, op, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	return a.toMedia(), nil
}

func handleSearchResponse(code int, body []byte, retryAfter int) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[searchResponse](types.SourceMAL, OpSearch, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, node := range resp.Data {
		results = append(results, *node.Node.toMedia())
	}
	return results, nil
}

func handleMangaResponse(code int, body []byte, op string, retryAfter int) (*types.Media, error) {
	m, err := helpers.DecodeJSON[manga](types.SourceMAL, op, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	return m.toMedia(), nil
}

// handleMangaListResponse serves manga search and manga ranking (same shape).
func handleMangaListResponse(code int, body []byte, op string, retryAfter int) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[mangaSearchResponse](types.SourceMAL, op, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, node := range resp.Data {
		results = append(results, *node.Node.toMedia())
	}
	return results, nil
}

func handleRankingResponse(code int, body []byte, retryAfter int) ([]types.Media, error) {
	resp, err := helpers.DecodeJSON[rankingResponse](types.SourceMAL, OpGetRanking, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Media, 0, len(resp.Data))
	for _, node := range resp.Data {
		results = append(results, *node.Node.toMedia())
	}
	return results, nil
}

func handleCharactersResponse(code int, body []byte, retryAfter int) ([]types.Character, error) {
	resp, err := helpers.DecodeJSON[charactersResponse](types.SourceMAL, OpGetAnimeCharacters, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	results := make([]types.Character, 0, len(resp.Data))
	for _, ac := range resp.Data {
		results = append(results, *ac.toCharacter())
	}
	return results, nil
}

func handleCharacterResponse(code int, body []byte, retryAfter int) (*types.Character, error) {
	c, err := helpers.DecodeJSON[character](types.SourceMAL, OpGetCharacter, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	return c.toCharacter(), nil
}

func handlePersonResponse(code int, body []byte, retryAfter int) (*types.Person, error) {
	p, err := helpers.DecodeJSON[person](types.SourceMAL, OpGetPerson, code, retryAfter, body)
	if err != nil {
		return nil, err
	}
	return p.toPerson(), nil
}
