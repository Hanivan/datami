package anilist

import (
	"context"

	"github.com/Hanivan/datami/types"
)

// Non-goal methods (not used in mono-crawler catalog path). Stubbed with ErrNotImplemented.
func (c *Client) GetViewer(ctx context.Context) (*types.UserInfo, error) {
	return nil, &types.Error{Source: types.SourceAniList, Op: OpGetViewer, Err: types.ErrNotImplemented}
}

func (c *Client) GetMediaListEntry(ctx context.Context, mediaID, userID int) (*types.MediaListEntry, error) {
	return nil, &types.Error{Source: types.SourceAniList, Op: OpGetMediaListEntry, Err: types.ErrNotImplemented}
}

func (c *Client) GetMediaListCollection(ctx context.Context, rt string, userID int) (*types.MediaListCollection, error) {
	return nil, &types.Error{Source: types.SourceAniList, Op: OpGetMediaListCollection, Err: types.ErrNotImplemented}
}

func (c *Client) SaveMediaListEntry(ctx context.Context, entry types.MediaListEntry) error {
	return &types.Error{Source: types.SourceAniList, Op: OpSaveMediaListEntry, Err: types.ErrNotImplemented}
}

func (c *Client) DeleteMediaListEntry(ctx context.Context, listEntryID int) error {
	return &types.Error{Source: types.SourceAniList, Op: OpDeleteMediaListEntry, Err: types.ErrNotImplemented}
}
