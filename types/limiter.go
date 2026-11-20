package types

import "context"

// Limiter gates outgoing requests. *rate.Limiter satisfies it, so several
// clients can share one budget.
type Limiter interface {
	Wait(ctx context.Context) error
}
