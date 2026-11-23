package anilist

import (
	"context"

	"github.com/Hanivan/datami/internal/helpers"
	"github.com/Hanivan/datami/types"
)

// Non-goal methods (not used in mono-crawler catalog path). Stubbed with ErrNotImplemented.
func (c *Client) GetViewer(ctx context.Context) (*types.UserInfo, error) {
	return nil, helpers.Wrap(types.SourceAniList, OpGetViewer, types.ErrNotImplemented)
}

func (c *Client) GetMediaListEntry(ctx context.Context, mediaID, userID int) (*types.MediaListEntry, error) {
	return nil, helpers.Wrap(types.SourceAniList, OpGetMediaListEntry, types.ErrNotImplemented)
}

func (c *Client) GetMediaListCollection(ctx context.Context, rt string, userID int) (*types.MediaListCollection, error) {
	return nil, helpers.Wrap(types.SourceAniList, OpGetMediaListCollection, types.ErrNotImplemented)
}

func (c *Client) SaveMediaListEntry(ctx context.Context, entry types.MediaListEntry) error {
	return helpers.Wrap(types.SourceAniList, OpSaveMediaListEntry, types.ErrNotImplemented)
}

func (c *Client) DeleteMediaListEntry(ctx context.Context, listEntryID int) error {
	return helpers.Wrap(types.SourceAniList, OpDeleteMediaListEntry, types.ErrNotImplemented)
}
