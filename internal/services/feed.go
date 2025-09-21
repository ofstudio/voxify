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
func NewFeedService(cfg *config.Settings, log *slog.Logger, s Store) *FeedService {
	return &FeedService{
		cfg:   cfg,
		log:   log,
		store: s,
	}
}

// Init checks the service dependencies and prepares the environment.
func (s *FeedService) Init(_ context.Context) error {
	return nil
}

// Build implements Feeder interface to generate RSS feed from all episodes.
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

	return feedcast.NewFeed(feedcast.FeedData{
		Title:       s.cfg.FeedTitle,
		Description: s.cfg.FeedDescription,
		Image:       s.cfg.FeedImage,
		Language:    s.cfg.FeedLanguage,
		Explicit:    explicit,
		Categories:  s.getCategories(),
	}).
		WithLink(s.cfg.FeedLink).
		WithItunesTitle(s.cfg.FeedTitle).
		WithItunesSummary(s.cfg.FeedDescription).
		WithItunesKeywords(s.cfg.FeedKeywords).
		WithAuthor(s.cfg.FeedAuthor).
		WithLastBuildDate(now).
		WithGenerator(s.getGenerator())
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

// getCategories converts a list of config.Settings categories to entities.FeedCategory.
func (s *FeedService) getCategories() []entities.FeedCategory {
	var categories []entities.FeedCategory
	if len(s.cfg.FeedCategories) > 0 {
		categories = append(categories, entities.FeedCategory{
			Text:          s.cfg.FeedCategories[0],
			Subcategories: s.cfg.FeedCategories[1:],
		})
	}
	if len(s.cfg.FeedCategories2) > 0 {
		categories = append(categories, entities.FeedCategory{
			Text:          s.cfg.FeedCategories2[0],
			Subcategories: s.cfg.FeedCategories2[1:],
		})
	}

	if len(s.cfg.FeedCategories3) > 0 {
		categories = append(categories, entities.FeedCategory{
			Text:          s.cfg.FeedCategories3[0],
			Subcategories: s.cfg.FeedCategories3[1:],
		})
	}

	return categories
}

// getGenerator returns the feed generator string.
func (s *FeedService) getGenerator() string {
	return "Voxify " + config.Version() + " (github.com/ofstudio/voxify)"
}

func (s *FeedService) Feed(ctx context.Context) (*entities.Feed, error) {
	var pubDate time.Time

	// Count episodes
	count, err := s.store.EpisodeCountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count episodes: %w", err)
	}

	// If there are episodes, get the last published date
	if count > 0 {
		if pubDate, err = s.store.EpisodeGetLastTime(ctx); err != nil {
			return nil, fmt.Errorf("failed to get last episode time: %w", err)
		}
	}

	return &entities.Feed{
		Title:         s.cfg.FeedTitle,
		Description:   s.cfg.FeedDescription,
		Summary:       s.cfg.FeedDescription, // For now, using description as summary
		Language:      s.cfg.FeedLanguage,
		Categories:    s.getCategories(),
		Keywords:      s.cfg.FeedKeywords,
		Author:        s.cfg.FeedAuthor,
		Owner:         nil, // Owner not implemented yet
		Copyright:     "",  // Copyright not implemented yet
		Explicit:      s.cfg.FeedIsExplicit,
		FeedType:      entities.FeedTypeNotSet, // Feed type not implemented yet
		FeedCompleted: false,                   // Feed completed feature not implemented yet
		FeedBlocked:   false,                   // Feed blocked feature not implemented yet
		WebsiteLink:   s.cfg.FeedLink,
		RSSLink:       s.cfg.PublicUrl.JoinPath(s.cfg.FeedFileName).String(),
		ImageUrl:      s.cfg.FeedImage,
		Generator:     s.getGenerator(),
		PubDate:       pubDate, // Zero time if no episodes
		EpisodeCount:  count,
	}, nil
}
