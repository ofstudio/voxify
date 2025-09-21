package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/mocks"
)

// TestFeedServiceSuite is a test suite for FeedService
type TestFeedServiceSuite struct {
	suite.Suite
	ctx       context.Context
	cfg       *config.Settings
	log       *slog.Logger
	mockStore *mocks.MockStore
	service   *FeedService
	tempDir   string
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestFeedServiceSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.log = slog.Default()

	// Create temporary directory for testing
	var err error
	suite.tempDir, err = os.MkdirTemp("", "voxify_test_feed_*")
	suite.Require().NoError(err)

	publicUrl, err := url.Parse("https://test.example.com/public")
	suite.Require().NoError(err)
	suite.cfg = &config.Settings{
		PublicDir:       suite.tempDir,
		PublicUrl:       *publicUrl,
		FeedFileName:    "feed.xml",
		FeedTitle:       "Test Podcast",
		FeedDescription: "Test podcast description",
		FeedLink:        "https://test.example.com",
		FeedImage:       "https://test.example.com/cover.jpg",
		FeedLanguage:    "en",
		FeedCategories:  []string{"Technology", "Podcasts"},
		FeedAuthor:      "Test Author",
		FeedIsExplicit:  false,
	}
}

// TearDownSuite is called once after the entire test suite runs
func (suite *TestFeedServiceSuite) TearDownSuite() {
	_ = os.RemoveAll(suite.tempDir)
}

// SetupSubTest is called before each subtest
func (suite *TestFeedServiceSuite) SetupSubTest() {
	suite.mockStore = mocks.NewMockStore(suite.T())
	suite.service = NewFeedService(suite.cfg, suite.log, suite.mockStore)
}

// TestNewFeedService tests the constructor
func (suite *TestFeedServiceSuite) TestNewFeedService() {
	// Act
	service := NewFeedService(suite.cfg, suite.log, suite.mockStore)

	// Assert
	suite.NotNil(service)
	suite.Equal(suite.cfg, service.cfg)
	suite.Equal(suite.mockStore, service.store)
	suite.NotNil(service.log)
}

// TestBuild tests the Build method
func (suite *TestFeedServiceSuite) TestBuild() {
	suite.Run("Success_NoEpisodes", func() {
		// Arrange
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return([]*entities.Episode{}, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, ErrEmptyFeed))
		// No feed file should be created when there are no episodes
	})

	suite.Run("Success_WithEpisodes", func() {
		// Arrange
		now := time.Now()
		episodes := []*entities.Episode{
			{
				ID:            1,
				OriginalURL:   "https://example.com/episode1",
				Title:         "Test Episode 1",
				Description:   "Description 1",
				CanonicalURL:  "https://example.com/episode1",
				CreatedAt:     now,
				MediaFile:     "episode1.mp3",
				MediaSize:     1024000,
				MediaType:     "audio/mpeg",
				MediaDuration: 3600,
				ThumbnailFile: "thumb1.jpg",
			},
			{
				ID:            2,
				OriginalURL:   "https://example.com/episode2",
				Title:         "Test Episode 2",
				Description:   "Description 2",
				CanonicalURL:  "https://example.com/episode2",
				CreatedAt:     now,
				MediaFile:     "episode2.mp3",
				MediaSize:     512000,
				MediaType:     "audio/mpeg",
				MediaDuration: 1800,
				ThumbnailFile: "",
			},
		}
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return(episodes, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.NoError(err)

		// Verify feed file was created
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		suite.FileExists(feedPath)

		// Verify feed content contains episodes
		content, err := os.ReadFile(feedPath)
		suite.NoError(err)
		feedContent := string(content)
		suite.Contains(feedContent, "Test Episode 1")
		suite.Contains(feedContent, "Test Episode 2")
		suite.Contains(feedContent, suite.cfg.FeedTitle)
		suite.Contains(feedContent, suite.cfg.FeedDescription)
	})

	suite.Run("Error_StoreFailure", func() {
		// Arrange
		expectedErr := errors.New("store error")
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return(nil, expectedErr)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to list all episodes")
		suite.Contains(err.Error(), "store error")
	})

	suite.Run("Error_InvalidPublicDir", func() {
		// Arrange
		now := time.Now()
		episodes := []*entities.Episode{
			{
				ID:            1,
				OriginalURL:   "https://example.com/episode1",
				Title:         "Test Episode 1",
				Description:   "Description 1",
				CanonicalURL:  "https://example.com/episode1",
				CreatedAt:     now,
				MediaFile:     "episode1.mp3",
				MediaSize:     1024000,
				MediaType:     "audio/mpeg",
				MediaDuration: 3600,
			},
		}
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return(episodes, nil)

		// Create service with invalid public directory
		invalidCfg := *suite.cfg
		invalidCfg.PublicDir = "/invalid/path/that/does/not/exist"
		service := NewFeedService(&invalidCfg, suite.log, suite.mockStore)

		// Act
		err := service.Build(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to save feed to file")
		suite.Contains(err.Error(), "failed to create feed file")
		suite.Contains(err.Error(), "no such file or directory")
	})

	suite.Run("Error_ContextCancelled", func() {
		// Arrange
		cancelledCtx, cancel := context.WithCancel(suite.ctx)
		cancel() // Cancel immediately

		suite.mockStore.On("EpisodeListAll", cancelledCtx).Return(nil, context.Canceled)

		// Act
		err := suite.service.Build(cancelledCtx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to list all episodes")
		suite.ErrorIs(err, context.Canceled)
	})
}

// TestBuild_EdgeCases tests edge cases for Build method
func (suite *TestFeedServiceSuite) TestBuild_EdgeCases() {
	suite.Run("EmptyConfig", func() {
		// Arrange
		emptyCfg := &config.Settings{}
		service := NewFeedService(emptyCfg, suite.log, suite.mockStore)
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return([]*entities.Episode{}, nil)

		// Act
		err := service.Build(suite.ctx)

		// Assert
		suite.Error(err) // Should fail due to ErrEmptyFeed
		suite.True(errors.Is(err, ErrEmptyFeed))
	})

	suite.Run("LargeNumberOfEpisodes", func() {
		// Arrange - create many episodes to test performance
		now := time.Now()
		episodes := make([]*entities.Episode, 100)
		for i := 0; i < 100; i++ {
			episodes[i] = &entities.Episode{
				ID:            int64(i + 1),
				OriginalURL:   fmt.Sprintf("https://example.com/episode%d", i),
				Title:         fmt.Sprintf("Test Episode %d", i),
				Description:   fmt.Sprintf("Description %d", i),
				CanonicalURL:  fmt.Sprintf("https://example.com/episode%d", i),
				CreatedAt:     now,
				MediaFile:     fmt.Sprintf("episode%d.mp3", i),
				MediaSize:     1024000,
				MediaType:     "audio/mpeg",
				MediaDuration: 3600,
			}
		}
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return(episodes, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.NoError(err)

		// Verify feed file was created and contains all episodes
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		suite.FileExists(feedPath)
	})

	suite.Run("EpisodeWithSpecialCharacters", func() {
		// Arrange
		now := time.Now()
		episodes := []*entities.Episode{
			{
				ID:            1,
				OriginalURL:   "https://example.com/episode1",
				Title:         "Test Episode with <special> & characters \"quotes\"",
				Description:   "Description with <XML> & entities",
				CanonicalURL:  "https://example.com/episode1",
				CreatedAt:     now,
				MediaFile:     "episode1.mp3",
				MediaSize:     1024000,
				MediaType:     "audio/mpeg",
				MediaDuration: 3600,
			},
		}
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return(episodes, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.NoError(err)

		// Verify feed file was created
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		suite.FileExists(feedPath)
	})

	suite.Run("SupportedMediaFormats", func() {
		// Arrange - test both supported formats
		now := time.Now()
		episodes := []*entities.Episode{
			{
				ID:            1,
				OriginalURL:   "https://example.com/episode1",
				Title:         "DownloadMp3 Episode",
				Description:   "DownloadMp3 Description",
				CanonicalURL:  "https://example.com/episode1",
				CreatedAt:     now,
				MediaFile:     "episode1.mp3",
				MediaSize:     1024000,
				MediaType:     "audio/mpeg",
				MediaDuration: 3600,
			},
			{
				ID:            2,
				OriginalURL:   "https://example.com/episode2",
				Title:         "M4A Episode",
				Description:   "M4A Description",
				CanonicalURL:  "https://example.com/episode2",
				CreatedAt:     now,
				MediaFile:     "episode2.m4a",
				MediaSize:     512000,
				MediaType:     "audio/mpeg",
				MediaDuration: 1800,
			},
		}
		suite.mockStore.On("EpisodeListAll", suite.ctx).Return(episodes, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.NoError(err)

		// Verify feed file was created
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		suite.FileExists(feedPath)
	})
}

// TestFeed method
func (suite *TestFeedServiceSuite) TestFeed() {
	suite.Run("WithEpisodes", func() {
		// Arrange
		now := time.Now().UTC()
		suite.mockStore.On("EpisodeCountAll", suite.ctx).Return(2, nil)
		suite.mockStore.On("EpisodeGetLastTime", suite.ctx).Return(now, nil)

		// Act
		feed, err := suite.service.Feed(suite.ctx)

		// Assert
		suite.Require().NoError(err)
		suite.Equal(2, feed.EpisodeCount)
		suite.True(feed.PubDate.Equal(now))
		suite.Equal(suite.cfg.FeedTitle, feed.Title)
		suite.Equal(suite.cfg.FeedDescription, feed.Description)
		suite.Equal(suite.cfg.PublicUrl.JoinPath(suite.cfg.FeedFileName).String(), feed.RSSLink)
	})

	suite.Run("ZeroEpisodes", func() {
		// Arrange
		suite.mockStore.On("EpisodeCountAll", suite.ctx).Return(0, nil)

		// Act
		feed, err := suite.service.Feed(suite.ctx)

		// Assert
		suite.Require().NoError(err)
		suite.Equal(0, feed.EpisodeCount)
		suite.True(feed.PubDate.IsZero())
	})

	suite.Run("CountError", func() {
		// Arrange
		expectedErr := errors.New("count failed")
		suite.mockStore.On("EpisodeCountAll", suite.ctx).Return(0, expectedErr)

		// Act
		_, err := suite.service.Feed(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to count all episodes")
		suite.ErrorIs(err, expectedErr)
	})

	suite.Run("LastTimeError", func() {
		// Arrange
		expectedErr := errors.New("last failed")
		suite.mockStore.On("EpisodeCountAll", suite.ctx).Return(1, nil)
		suite.mockStore.On("EpisodeGetLastTime", suite.ctx).Return(time.Time{}, expectedErr)

		// Act
		_, err := suite.service.Feed(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to get last episode time")
		suite.ErrorIs(err, expectedErr)
	})
}

// TestFeedService runs the test suite
func TestFeedService(t *testing.T) {
	suite.Run(t, new(TestFeedServiceSuite))
}
