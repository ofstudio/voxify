package telegram

import (
	"context"

	"github.com/ofstudio/voxify/internal/entities"
)

type Processor interface {
	In() chan<- *entities.Request
}

type Builder interface {
	Build(ctx context.Context) error
}

type Notifier interface {
	Notify() <-chan *entities.Process
}
