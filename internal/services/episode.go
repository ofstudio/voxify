package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/pkg/files"
)

type EpisodeService struct {
	cfg       config.Settings
	log       *slog.Logger
	store     domain.Store
	platforms []domain.Platform
}

// NewEpisodeService creates a new EpisodeService instance.
func NewEpisodeService(cfg config.Settings, log *slog.Logger, s domain.Store, p ...domain.Platform) *EpisodeService {
	return &EpisodeService{
		cfg:       cfg,
		log:       log,
		store:     s,
		platforms: p,
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

	if err := s.enforceFeedMaxEpisodes(ctx); err != nil {
		return fmt.Errorf("failed to enforce feed episode limit: %w", err)
	}

	return nil
}

func (s *EpisodeService) Validate(ctx context.Context, req domain.DownloadRequest) error {
	// validate request
	if err := s.validateRequest(ctx, req); err != nil {
		return err
	}
	// find platform
	if s.findPlatform(req) == nil {
		return domain.ErrNoMatchingPlatform
	}

	return nil
}

// Download downloads an Episode from the given URL using the appropriate platform.
func (s *EpisodeService) Download(ctx context.Context, req domain.DownloadRequest) (*domain.Episode, error) {
	if err := s.validateRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	platform := s.findPlatform(req)
	if platform == nil {
		return nil, domain.ErrNoMatchingPlatform
	}

	s.log.Info("[episode service] downloading episode",
		"platform", platform.ID(), "request", req.LogValue())
	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.DownloadTimeout)
	defer cancel()

	episode, err := platform.Download(ctxTimeout, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrDownloadFailed, err)
	}

	// Save Episode to store
	if err = s.store.EpisodeCreate(ctx, episode); err != nil {
		return nil, fmt.Errorf("failed to create episode: %w", err)
	}

	s.log.Info("[episode service] episode downloaded",
		"platform", platform.ID(), "request", req.LogValue(), "episode", episode.LogValue())

	if err = s.enforceFeedMaxEpisodes(ctx); err != nil {
		return nil, fmt.Errorf("failed to enforce feed episode limit: %w", err)
	}

	return episode, nil
}

func (s *EpisodeService) enforceFeedMaxEpisodes(ctx context.Context) error {
	count, err := s.store.EpisodeCount(ctx)
	if err != nil {
		return fmt.Errorf("failed to count episodes: %w", err)
	}

	if s.cfg.FeedMaxEpisodes <= 0 {
		s.log.Info("[episode service] feed episode limit disabled", "episodes_count", count)
		return nil
	}

	s.log.Info("[episode service] checking feed episode limit",
		"feed_max_episodes", s.cfg.FeedMaxEpisodes, "episodes_count", count)

	overflow := count - s.cfg.FeedMaxEpisodes
	if overflow <= 0 {
		return nil
	}

	episodes, err := s.store.EpisodeGetOldest(ctx, overflow)
	if err != nil {
		return fmt.Errorf("failed to get oldest episodes: %w", err)
	}

	for _, episode := range episodes {
		if err = s.store.EpisodeDelete(ctx, episode.ID); err != nil {
			return fmt.Errorf("failed to delete episode: %w", err)
		}
		if err = s.deleteEpisodeFiles(episode); err != nil {
			return fmt.Errorf("failed to delete episode files: %w", err)
		}
		s.log.Info("[episode service] old episode deleted",
			"feed_max_episodes", s.cfg.FeedMaxEpisodes, "episode", episode.LogValue())
	}

	return nil
}

func (s *EpisodeService) deleteEpisodeFiles(episode *domain.Episode) error {
	for _, filename := range []string{episode.MediaFile, episode.ThumbnailFile} {
		if filename == "" {
			continue
		}
		if err := s.deletePublicFile(filename); err != nil {
			return err
		}
	}
	return nil
}

func (s *EpisodeService) deletePublicFile(filename string) error {
	clean := filepath.Clean(filename)
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return fmt.Errorf("unsafe public file path: %s", filename)
	}

	path := filepath.Join(s.cfg.PublicDir, clean)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove public file %s: %w", filename, err)
	}
	return nil
}

func (s *EpisodeService) findPlatform(req domain.DownloadRequest) domain.Platform {
	for _, p := range s.platforms {
		if p.Match(req) {
			return p
		}
	}
	return nil
}

// validateRequest validates the download request.
func (s *EpisodeService) validateRequest(ctx context.Context, req domain.DownloadRequest) error {
	if err := s.validateUrl(req.Url); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidUrl, err)
	}
	if err := s.validateDownloadFormat(req.DownloadFormat); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidFormat, err)
	}
	if err := s.validateDownloadQuality(req.DownloadQuality); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrDownloadQuality, err)
	}

	//check if episode already exists
	existing, err := s.store.EpisodeGetByUrl(ctx, req.Url)
	if err != nil {
		return fmt.Errorf("failed to get existing episodes: %w", err)
	}
	if len(existing) > 0 {
		return domain.ErrDownloadExists
	}
	return nil
}

func (s *EpisodeService) validateUrl(href string) error {
	u, err := url.Parse(href)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported url scheme: %s", u.Scheme)
	}
	return nil
}

func (s *EpisodeService) validateDownloadFormat(format domain.DownloadFormat) error {
	switch format {
	case domain.DownloadMp3:
		return nil
	case domain.DownloadM4a:
		return nil
	case "":
		return errors.New("download format not specified")
	default:
		return fmt.Errorf("unsupported download format: %s", format)
	}
}

var reSafeQuality = regexp.MustCompile(`^[0-9a-zA-Z-_]{1,32}$`)

func (s *EpisodeService) validateDownloadQuality(quality string) error {
	if quality == "" {
		return errors.New("download quality is empty")
	}
	if !reSafeQuality.MatchString(quality) {
		return fmt.Errorf("unsupported download quality: %s", quality)
	}
	return nil
}
