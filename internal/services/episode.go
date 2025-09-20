package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/pkg/files"
)

type EpisodeService struct {
	cfg       *config.Settings
	log       *slog.Logger
	store     Store
	platforms []Platform
}

// NewEpisodeService creates a new EpisodeService instance.
func NewEpisodeService(cfg *config.Settings, log *slog.Logger, store Store, platforms ...Platform) *EpisodeService {
	return &EpisodeService{
		cfg:       cfg,
		log:       log,
		store:     store,
		platforms: platforms,
	}
}

// Init checks the service dependencies and prepares the environment.
func (s *EpisodeService) Init(ctx context.Context) error {

	// init platforms
	for _, p := range s.platforms {
		if err := p.Init(ctx); err != nil {
			return fmt.Errorf("platform check failed: %w", err)
		}
	}

	// check public directory exists
	if err := files.IsDir(s.cfg.PublicDir); err != nil {
		return fmt.Errorf("public directory check failed: %w", err)
	}

	// check download directory exists
	if err := files.IsDir(s.cfg.DownloadDir); err != nil {
		return fmt.Errorf("download directory check failed: %w", err)
	}

	// cleanup download directory on startup
	if err := files.CleanDir(s.cfg.DownloadDir); err != nil {
		return fmt.Errorf("failed to clean download directory: %w", err)
	}

	return nil
}

// Download downloads an episode from the given URL using the appropriate platform.
func (s *EpisodeService) Download(ctx context.Context, req entities.Request) (*entities.Episode, error) {
	if req.DownloadFormat == "" {
		req.DownloadFormat = s.cfg.DownloadFormat
	}
	if req.DownloadQuality == "" {
		req.DownloadQuality = s.cfg.DownloadQuality
	}
	platform := s.findPlatform(req.Url)
	if platform == nil {
		return nil, ErrNoMatchingPlatform
	}

	s.log.Info("[episode service] downloading episode",
		"platform", platform.ID(), "request", req.LogValue())
	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.DownloadTimeout)
	defer cancel()

	episode, err := platform.Download(ctxTimeout, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	// Save episode to store
	if err = s.store.EpisodeCreate(ctx, episode); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEpisodeCreate, err)
	}

	s.log.Info("[episode service] episode downloaded",
		"platform", platform.ID(), "request", req.LogValue(), "episode", episode.LogValue())
	return episode, nil
}

func (s *EpisodeService) findPlatform(url string) Platform {
	for _, p := range s.platforms {
		if p.Match(url) {
			return p
		}
	}
	return nil
}
