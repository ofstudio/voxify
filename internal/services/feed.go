package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/templates"
	"github.com/ofstudio/voxify/pkg/feedcast"
)

// FeedService builds RSS podcast FeedInfo from episodes.
type FeedService struct {
	cfg   config.Settings
	log   *slog.Logger
	store domain.Store
}

// NewFeedService creates a new FeedService instance.
func NewFeedService(cfg config.Settings, log *slog.Logger, s domain.Store) *FeedService {
	return &FeedService{
		cfg:   cfg,
		log:   log,
		store: s,
	}
}

// Init checks the service dependencies and prepares the environment.
func (s *FeedService) Init(ctx context.Context) error {
	if templates.LandingTemplate == nil {
		return errors.New("landing page template is not initialized")
	}
	if err := s.Build(ctx); err != nil {
		return err
	}
	return nil
}

// Build implements Feeder interface to generate RSS FeedInfo from all episodes and update landing page
func (s *FeedService) Build(ctx context.Context) error {
	// Get all episodes from store
	episodes, err := s.store.EpisodeGet(ctx, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}

	// Build landing page
	if err = s.buildLanding(ctx, episodes); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrBuildLandingFailed, err)
	}

	if len(episodes) == 0 {
		s.log.Info("[feed service] no episodes found, skipping feed build")
		return nil
	}

	// Build RSS feed
	if err = s.buildRss(ctx, episodes); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrBuildFeedFailed, err)
	}

	return nil
}

func (s *FeedService) Info(ctx context.Context) (*domain.FeedInfo, error) {
	var pubDate time.Time

	// Count episodes
	count, err := s.store.EpisodeCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count episodes: %w", err)
	}

	// If there are episodes, get the last published
	if count > 0 {
		var recents []*domain.Episode
		recents, err = s.store.EpisodeGet(ctx, 1, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get most recent episode: %w", err)
		}
		if len(recents) > 0 {
			pubDate = recents[0].CreatedAt
		}
	}

	return s.feedInfo(count, pubDate), nil
}

func (s *FeedService) feedInfo(episodeCount int, pubDate time.Time) *domain.FeedInfo {
	return &domain.FeedInfo{
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
		FeedType:      domain.FeedTypeNotSet, // Feed type not implemented yet
		FeedCompleted: false,                 // Feed completed feature not implemented yet
		FeedBlocked:   false,                 // Feed blocked feature not implemented yet
		WebsiteLink:   s.cfg.FeedLink,
		RSSLink:       s.cfg.PublicUrl.JoinPath(s.cfg.FeedFileName).String(),
		ImageUrl:      s.cfg.FeedImage,
		Generator:     s.getGenerator(),
		PubDate:       pubDate, // Zero time if no episodes
		EpisodeCount:  episodeCount,
	}
}

func (s *FeedService) buildRss(ctx context.Context, episodes []*domain.Episode) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	s.log.Info("[feed service] building podcast feed")

	// Create podcast feed
	feed := s.createFeed()

	// Add publication date if there are episodes
	if len(episodes) > 0 {
		feed.WithPubDate(episodes[0].CreatedAt)
	}

	// Add episodes to feed
	for _, episode := range episodes {
		feed.AddItem(s.createItem(episode))
	}

	// Write feed to file
	if err := s.saveFeed(feed); err != nil {
		return fmt.Errorf("failed to save feed file: %w", err)
	}

	s.log.Info("[feed service] podcast feed built", "episodes_count", len(episodes))

	return nil
}

// createFeed creates and configures the main podcast FeedInfo.
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

// createItem creates a FeedInfo item from an Episode entity.
func (s *FeedService) createItem(episode *domain.Episode) *feedcast.Item {
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

// saveFeed writes the RSS FeedInfo to the configured file path.
func (s *FeedService) saveFeed(feed *feedcast.Feed) error {

	feedPath := filepath.Join(s.cfg.PublicDir, s.cfg.FeedFileName)
	file, tmpPath, err := s.createTempFile(feedPath)
	if err != nil {
		return fmt.Errorf("failed to create feed file: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer os.Remove(tmpPath)

	// Write RSS feed to file
	if err = feed.Encode(file); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to encode feed to file: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("failed to close feed file: %w", err)
	}
	if err = os.Rename(tmpPath, feedPath); err != nil {
		return fmt.Errorf("failed to replace feed file: %w", err)
	}

	return nil
}

// getCategories converts a list of config.Settings categories to entities.FeedCategory.
func (s *FeedService) getCategories() []domain.FeedCategory {
	var categories []domain.FeedCategory
	if len(s.cfg.FeedCategories) > 0 {
		categories = append(categories, domain.FeedCategory{
			Text:          s.cfg.FeedCategories[0],
			Subcategories: s.cfg.FeedCategories[1:],
		})
	}
	if len(s.cfg.FeedCategories2) > 0 {
		categories = append(categories, domain.FeedCategory{
			Text:          s.cfg.FeedCategories2[0],
			Subcategories: s.cfg.FeedCategories2[1:],
		})
	}

	if len(s.cfg.FeedCategories3) > 0 {
		categories = append(categories, domain.FeedCategory{
			Text:          s.cfg.FeedCategories3[0],
			Subcategories: s.cfg.FeedCategories3[1:],
		})
	}

	return categories
}

// getGenerator returns the FeedInfo generator string.
func (s *FeedService) getGenerator() string {
	return "Voxify " + config.Version() + " (github.com/ofstudio/voxify)"
}

// buildLanding generates or updates the landing page HTML file.
func (s *FeedService) buildLanding(ctx context.Context, episodes []*domain.Episode) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if templates.LandingTemplate == nil {
		return errors.New("landing page template is not initialized")

	}
	s.log.Info("[feed service] building landing page")

	var pubDate time.Time
	if len(episodes) > 0 {
		pubDate = episodes[0].CreatedAt
	}

	// Get feed info
	feed := s.feedInfo(len(episodes), pubDate)

	// Prepare data for template
	data := struct {
		Feed     *domain.FeedInfo
		Episodes []*domain.Episode
	}{
		Feed:     feed,
		Episodes: episodes,
	}

	landingPath := filepath.Join(s.cfg.PublicDir, "index.html")
	file, tmpPath, err := s.createTempFile(landingPath)
	if err != nil {
		return fmt.Errorf("failed to create landing page file: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer os.Remove(tmpPath)

	// Execute template and write to file
	if err = templates.LandingTemplate.Execute(file, data); err != nil {
		_ = file.Close()
		return fmt.Errorf("failed to execute landing page template: %w", err)
	}
	if err = file.Close(); err != nil {
		return fmt.Errorf("failed to close landing page file: %w", err)
	}
	if err = os.Rename(tmpPath, landingPath); err != nil {
		return fmt.Errorf("failed to replace landing page file: %w", err)
	}

	s.log.Info("[feed service] landing page built", "episodes_count", len(episodes))
	return nil
}

func (s *FeedService) createTempFile(path string) (*os.File, string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return nil, "", err
	}
	return file, file.Name(), nil
}
