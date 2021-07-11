package ledger

import "context"

type HealthRepository interface {
	Cheak(ctx context.Context) error
}
