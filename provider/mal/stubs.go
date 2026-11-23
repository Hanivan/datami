package mal

import (
	"context"
	"net/url"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// Non-goal methods (not used in mono-crawler catalog path). Stubbed with ErrNotImplemented.
func (c *Client) GetForumBoards(ctx context.Context) ([]types.ForumBoard, error) {
	return nil, helpers.Wrap(types.SourceMAL, OpGetForumBoards, types.ErrNotImplemented)
}

func (c *Client) GetForumTopic(ctx context.Context, topicID string, limit int) ([]types.ForumPost, error) {
	return nil, helpers.Wrap(types.SourceMAL, OpGetForumTopic, types.ErrNotImplemented)
}

func (c *Client) SearchForumTopics(ctx context.Context, q string, limit int) ([]types.ForumTopic, error) {
	return nil, helpers.Wrap(types.SourceMAL, OpSearchForumTopics, types.ErrNotImplemented)
}

func (c *Client) GetUserAnimeList(ctx context.Context, userName string, limit int) ([]types.Media, error) {
	return nil, helpers.Wrap(types.SourceMAL, OpGetUserAnimeList, types.ErrNotImplemented)
}

func (c *Client) GetUserMangaList(ctx context.Context, userName string, limit int) ([]types.Media, error) {
	return nil, helpers.Wrap(types.SourceMAL, OpGetUserMangaList, types.ErrNotImplemented)
}

func (c *Client) GetUserInfo(ctx context.Context) ([]types.UserInfo, error) {
	return nil, helpers.Wrap(types.SourceMAL, OpGetUserInfo, types.ErrNotImplemented)
}

func (c *Client) UpdateAnimeListStatus(ctx context.Context, animeID int, form url.Values) error {
	return helpers.Wrap(types.SourceMAL, OpUpdateAnimeListStatus, types.ErrNotImplemented)
}

func (c *Client) DeleteAnimeListItem(ctx context.Context, animeID int) error {
	return helpers.Wrap(types.SourceMAL, OpDeleteAnimeListItem, types.ErrNotImplemented)
}
