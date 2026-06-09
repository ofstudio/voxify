package services

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/templates"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/mocks"
)

// TestFeedService runs the test suite
func TestFeedService(t *testing.T) {
	suite.Run(t, new(TestFeedServiceSuite))
}

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
	suite.Require().NoError(os.RemoveAll(suite.tempDir))
	suite.Require().NoError(os.MkdirAll(suite.tempDir, 0755))
	suite.mockStore = mocks.NewMockStore(suite.T())
	suite.service = NewFeedService(*suite.cfg, suite.log, suite.mockStore)
	suite.NoError(templates.Init(suite.ctx))
}

func (suite *TestFeedServiceSuite) fileModeIs(path string, expected os.FileMode) {
	info, err := os.Stat(path)
	suite.Require().NoError(err)
	suite.Equal(expected, info.Mode().Perm())
}

// TestNewFeedService tests the constructor
func (suite *TestFeedServiceSuite) TestNewFeedService() {
	// Act
	service := NewFeedService(*suite.cfg, suite.log, suite.mockStore)

	// Assert
	suite.NotNil(service)
	suite.Equal(*suite.cfg, service.cfg)
	suite.Equal(suite.mockStore, service.store)
	suite.NotNil(service.log)
}

// TestBuild tests the Build method
func (suite *TestFeedServiceSuite) TestBuild() {
	suite.Run("Success_NoEpisodes", func() {
		// Arrange
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return([]*domain.Episode{}, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		// 1) Landing page should be created
		suite.Require().NoError(err)
		landingPath := filepath.Join(suite.cfg.PublicDir, "index.html")
		suite.FileExists(landingPath)
		suite.fileModeIs(landingPath, 0644)

		// 2) Landing page must contain FeedTitle and FeedDescription
		content, readErr := os.ReadFile(landingPath)
		suite.NoError(readErr)
		page := string(content)
		suite.Contains(page, suite.cfg.FeedTitle)
		suite.Contains(page, suite.cfg.FeedDescription)

		// 3) Landing page must NOT contain the feed filename
		suite.NotContains(page, suite.cfg.FeedFileName)

		// When no episodes exist, RSS elements should not be present in landing page
		suite.NotContains(page, `<link rel="alternate" type="application/rss+xml"`, "Should not contain RSS link tag in head when no episodes")
		suite.NotContains(page, "RSS feed", "Should not show RSS feed button when no episodes")

		// 4) RSS feed file must NOT be created
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		if _, statErr := os.Stat(feedPath); statErr == nil {
			suite.Failf("feed file should not exist", "unexpected feed file present: %s", feedPath)
		} else {
			suite.True(os.IsNotExist(statErr))
		}
	})

	suite.Run("Success_WithEpisodes", func() {
		// Arrange
		now := time.Now()
		episodes := []*domain.Episode{
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
		// Only ONE call to EpisodeGet at the beginning of Build()
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return(episodes, nil).Once()

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.NoError(err)

		// Verify feed file was created
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		suite.FileExists(feedPath)
		suite.fileModeIs(feedPath, 0644)

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
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return(nil, expectedErr)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to get episodes")
		suite.Contains(err.Error(), "store error")
	})

	suite.Run("Error_InvalidPublicDir", func() {
		// Arrange
		now := time.Now()
		episodes := []*domain.Episode{
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
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return(episodes, nil)

		// Create service with invalid public directory
		invalidCfg := *suite.cfg
		invalidCfg.PublicDir = "/invalid/path/that/does/not/exist"
		service := NewFeedService(invalidCfg, suite.log, suite.mockStore)

		// Act
		err := service.Build(suite.ctx)

		// Assert
		suite.Error(err)
		// Error should be wrapped with domain.ErrBuildFeedFailed and include save/create failure
		suite.Contains(err.Error(), "failed to build landing page")
		suite.Contains(err.Error(), "failed to create landing page file")
	})

	suite.Run("Error_RssBuildKeepsExistingFeed", func() {
		// Arrange
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		oldContent := []byte("existing feed")
		suite.Require().NoError(os.WriteFile(feedPath, oldContent, 0644))

		episodes := []*domain.Episode{
			{
				ID:            1,
				OriginalURL:   "https://example.com/episode1",
				Title:         "Test Episode 1",
				Description:   "Description 1",
				CanonicalURL:  "https://example.com/episode1",
				CreatedAt:     time.Now(),
				MediaFile:     "episode1.mp3",
				MediaSize:     0,
				MediaType:     "audio/mpeg",
				MediaDuration: 3600,
			},
		}
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return(episodes, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.Error(err)
		content, readErr := os.ReadFile(feedPath)
		suite.NoError(readErr)
		suite.Equal(oldContent, content)
	})

	suite.Run("Error_LandingBuildKeepsExistingPage", func() {
		// Arrange
		landingPath := filepath.Join(suite.cfg.PublicDir, "index.html")
		oldContent := []byte("existing landing")
		suite.Require().NoError(os.WriteFile(landingPath, oldContent, 0644))

		originalTemplate := templates.LandingTemplate
		defer func() { templates.LandingTemplate = originalTemplate }()
		templates.LandingTemplate = template.Must(template.New("broken").
			Option("missingkey=error").
			Parse(`{{.Missing}}`))

		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return([]*domain.Episode{}, nil)

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.Error(err)
		content, readErr := os.ReadFile(landingPath)
		suite.NoError(readErr)
		suite.Equal(oldContent, content)
	})

	suite.Run("Error_ContextCancelled", func() {
		// Arrange
		cancelledCtx, cancel := context.WithCancel(suite.ctx)
		cancel() // Cancel immediately

		suite.mockStore.On("EpisodeGet", cancelledCtx, 0, 0).Return(nil, context.Canceled)

		// Act
		err := suite.service.Build(cancelledCtx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to get episodes")
		suite.Contains(err.Error(), "context canceled")
	})
}

// TestBuild_EdgeCases tests edge cases for Build method
func (suite *TestFeedServiceSuite) TestBuild_EdgeCases() {
	suite.Run("LargeNumberOfEpisodes", func() {
		// Arrange - create many episodes to test performance
		now := time.Now()
		episodes := make([]*domain.Episode, 100)
		for i := 0; i < 100; i++ {
			episodes[i] = &domain.Episode{
				ID:            int64(i + 1),
				OriginalURL:   fmt.Sprintf("https://example.com/episode%d", i),
				Title:         fmt.Sprintf("Test Episode %d", i),
				Description:   fmt.Sprintf("Description %d", i),
				CanonicalURL:  fmt.Sprintf("https://example.com/episode%d", i),
				CreatedAt:     now,
				MediaFile:     fmt.Sprintf("Episode%d.mp3", i),
				MediaSize:     1024000,
				MediaType:     "audio/mpeg",
				MediaDuration: 3600,
			}
		}
		// Only ONE call to EpisodeGet at the beginning of Build()
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return(episodes, nil).Once()

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
		episodes := []*domain.Episode{
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
		// Only ONE call to EpisodeGet at the beginning of Build()
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return(episodes, nil).Once()

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
		episodes := []*domain.Episode{
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
				MediaType:     domain.MediaM4a,
				MediaDuration: 1800,
			},
		}
		// Only ONE call to EpisodeGet at the beginning of Build()
		suite.mockStore.On("EpisodeGet", suite.ctx, 0, 0).Return(episodes, nil).Once()

		// Act
		err := suite.service.Build(suite.ctx)

		// Assert
		suite.NoError(err)

		// Verify feed file was created
		feedPath := filepath.Join(suite.cfg.PublicDir, suite.cfg.FeedFileName)
		suite.FileExists(feedPath)
		content, err := os.ReadFile(feedPath)
		suite.NoError(err)
		suite.Contains(string(content), `type="audio/mpeg"`)
		suite.Contains(string(content), `type="audio/x-m4a"`)
	})
}

// TestInfo method
func (suite *TestFeedServiceSuite) TestInfo() {
	suite.Run("WithEpisodes", func() {
		// Arrange
		suite.service.cfg.FeedMaxEpisodes = 50
		now := time.Now().UTC()
		recentEpisode := []*domain.Episode{
			{
				ID:        1,
				CreatedAt: now,
				Title:     "Recent Episode",
			},
		}

		suite.mockStore.On("EpisodeCount", suite.ctx).Return(2, nil)
		suite.mockStore.On("EpisodeGet", suite.ctx, 1, 0).Return(recentEpisode, nil)

		// Act
		feed, err := suite.service.Info(suite.ctx)

		// Assert
		suite.Require().NoError(err)
		suite.Equal(2, feed.EpisodeCount)
		suite.Equal(50, feed.FeedMaxEpisodes)
		suite.True(feed.PubDate.Equal(now))
		suite.Equal(suite.cfg.FeedTitle, feed.Title)
		suite.Equal(suite.cfg.FeedDescription, feed.Description)
		suite.Equal(suite.cfg.PublicUrl.JoinPath(suite.cfg.FeedFileName).String(), feed.RSSLink)
	})

	suite.Run("ZeroEpisodes", func() {
		// Arrange
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(0, nil)

		// Act
		feed, err := suite.service.Info(suite.ctx)

		// Assert
		suite.Require().NoError(err)
		suite.Equal(0, feed.EpisodeCount)
		suite.True(feed.PubDate.IsZero())
	})

	suite.Run("CountError", func() {
		// Arrange
		expectedErr := errors.New("count failed")
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(0, expectedErr)

		// Act
		_, err := suite.service.Info(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to count episodes")
		suite.ErrorIs(err, expectedErr)
	})

	suite.Run("EpisodeGetError", func() {
		// Arrange
		expectedErr := errors.New("episode get failed")
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(1, nil)
		suite.mockStore.On("EpisodeGet", suite.ctx, 1, 0).Return(nil, expectedErr)

		// Act
		_, err := suite.service.Info(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "failed to get most recent episode")
		suite.ErrorIs(err, expectedErr)
	})

	suite.Run("EmptyEpisodeGetResult", func() {
		// Arrange - count > 0 but EpisodeGet returns empty slice
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(1, nil)
		suite.mockStore.On("EpisodeGet", suite.ctx, 1, 0).Return([]*domain.Episode{}, nil)

		// Act
		feed, err := suite.service.Info(suite.ctx)

		// Assert
		suite.Require().NoError(err)
		suite.Equal(1, feed.EpisodeCount)
		suite.True(feed.PubDate.IsZero()) // Should be zero since no episodes returned
	})
}
