package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/platforms"
)

// Services holds all service instances.
type Services struct {
	Feed    FeedService
	Episode EpisodeService
}

// New creates a new Services instance.
func New(cfg config.Settings, log *slog.Logger, store domain.Store) *Services {
	ytDlp := platforms.NewYtDlpPlatform(cfg, log)
	feed := NewFeedService(cfg, log, store)
	episode := NewEpisodeService(cfg, log, store, ytDlp)

	return &Services{
		Feed:    *feed,
		Episode: *episode,
	}

}

// Init initializes all services
func (s Services) Init(ctx context.Context) error {
	// Initialize services
	if err := s.Episode.Init(ctx); err != nil {
		return fmt.Errorf("es service init failed: %w", err)
	}
	if err := s.Feed.Init(ctx); err != nil {
		return fmt.Errorf("feed service init failed: %w", err)
	}
	return nil
}
