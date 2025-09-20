package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/pkg/feedcast"
)

// FeedService builds RSS podcast feed from episodes.
type FeedService struct {
	cfg   *config.Settings
	log   *slog.Logger
	store Store
}

// NewFeedService creates a new FeedService instance.
func NewFeedService(cfg *config.Settings, log *slog.Logger, store Store) *FeedService {
	return &FeedService{
		cfg:   cfg,
		log:   log,
		store: store,
	}
}

// Init checks the service dependencies and prepares the environment.
func (s *FeedService) Init(_ context.Context) error {
	return nil
}

// Build implements Builder interface to generate RSS feed from all episodes.
func (s *FeedService) Build(ctx context.Context) error {
	s.log.Info("[feed service] building podcast feed")

	// Get all episodes from store
	episodes, err := s.store.EpisodeListAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}
	if len(episodes) == 0 {
		return ErrEmptyFeed
	}

	// Create podcast feed
	feed := s.createFeed().WithPubDate(episodes[0].CreatedAt)

	// Add episodes to feed
	for _, episode := range episodes {
		feed.AddItem(s.createItem(episode))
	}

	// Write feed to file
	if err = s.saveFeed(feed); err != nil {
		return fmt.Errorf("failed to write feed: %w", err)
	}

	s.log.Info("[feed service] podcast feed built", "episodes_count", len(episodes))

	return nil
}

// createFeed creates and configures the main podcast feed.
func (s *FeedService) createFeed() *feedcast.Feed {
	now := time.Now()
	explicit := feedcast.ExplicitFalse
	if s.cfg.FeedIsExplicit {
		explicit = feedcast.ExplicitTrue
	}

	var categories []feedcast.Category
	if len(s.cfg.FeedCategories) > 0 {
		categories = append(categories, feedcast.NewCategory(s.cfg.FeedCategories[0], s.cfg.FeedCategories[1:]...))
	}
	if len(s.cfg.FeedCategories2) > 0 {
		categories = append(categories, feedcast.NewCategory(s.cfg.FeedCategories2[0], s.cfg.FeedCategories2[1:]...))
	}
	if len(s.cfg.FeedCategories3) > 0 {
		categories = append(categories, feedcast.NewCategory(s.cfg.FeedCategories3[0], s.cfg.FeedCategories3[1:]...))
	}

	return feedcast.NewFeed(feedcast.FeedData{
		Title:       s.cfg.FeedTitle,
		Description: s.cfg.FeedDescription,
		Image:       s.cfg.FeedImage,
		Language:    s.cfg.FeedLanguage,
		Explicit:    explicit,
		Categories:  categories,
	}).
		WithLink(s.cfg.FeedLink).
		WithItunesTitle(s.cfg.FeedTitle).
		WithItunesSummary(s.cfg.FeedDescription).
		WithAuthor(s.cfg.FeedAuthor).
		WithLastBuildDate(now).
		WithGenerator("Voxify " + config.Version() + " (github.com/ofstudio/voxify)")
}

// createItem creates a feed item from an episode entity.
func (s *FeedService) createItem(episode *entities.Episode) *feedcast.Item {

	// Create item
	mediaUrl := s.cfg.PublicUrl.JoinPath(episode.MediaFile).String()
	item := feedcast.NewItem(feedcast.ItemData{
		Title:     episode.Title,
		Enclosure: feedcast.NewEnclosure(mediaUrl, episode.MediaSize, episode.MediaType),
		Guid:      mediaUrl,
	}).
		WithItunesDuration(episode.MediaDuration).
		WithPubDate(episode.CreatedAt).
		WithDescription(episode.Description).
		WithItunesTitle(episode.Title).
		WithItunesSummary(episode.Description).
		WithLink(episode.CanonicalURL).
		WithItunesAuthor(episode.Author)

	if episode.ThumbnailFile != "" {
		thumbUrl := s.cfg.PublicUrl.JoinPath(episode.ThumbnailFile).String()
		item = item.WithItunesImage(thumbUrl)
	}

	return item
}

// saveFeed writes the RSS feed to the configured file path.
func (s *FeedService) saveFeed(feed *feedcast.Feed) error {

	// Create or overwrite feed file
	file, err := os.Create(filepath.Join(s.cfg.PublicDir, s.cfg.FeedFileName))
	if err != nil {
		return fmt.Errorf("failed to create feed file: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	// Write RSS feed to file
	if err = feed.Encode(file); err != nil {
		return fmt.Errorf("failed to encode feed: %w", err)
	}

	return nil
}
