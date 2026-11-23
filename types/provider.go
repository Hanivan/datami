package types

import "context"

// Provider is the lookup surface shared by every source client.
type Provider interface {
	Get(ctx context.Context, id string) (*Media, error)
	Search(ctx context.Context, query string, opts SearchOpts) ([]Media, error)
	GetByMalID(ctx context.Context, malID string) (*Media, error)
}
