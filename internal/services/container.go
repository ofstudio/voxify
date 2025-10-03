package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
)

// Container holds all service instances.
type Container struct {
	Feed    *FeedService
	Episode *EpisodeService
}

// New creates a new service Container instance.
func New(cfg config.Settings, log *slog.Logger, store domain.Store, platforms ...domain.Platform) *Container {
	return &Container{
		Feed:    NewFeedService(cfg, log, store),
		Episode: NewEpisodeService(cfg, log, store, platforms...),
	}

}

// Init initializes all services
func (c Container) Init(ctx context.Context) error {
	if err := c.Episode.Init(ctx); err != nil {
		return fmt.Errorf("episode service init failed: %w", err)
	}
	if err := c.Feed.Init(ctx); err != nil {
		return fmt.Errorf("feed service init failed: %w", err)
	}
	return nil
}
