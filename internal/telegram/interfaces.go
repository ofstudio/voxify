package telegram

import (
	"context"
)

type Builder interface {
	Build(ctx context.Context) error
}
