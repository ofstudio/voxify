package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// TestEpisodeService runs the test suite
func TestEpisodeService(t *testing.T) {
	suite.Run(t, new(TestEpisodeServiceSuite))
}

// TestEpisodeServiceSuite is a test suite for EpisodeService
type TestEpisodeServiceSuite struct {
	suite.Suite
	ctx          context.Context
	cfg          config.Settings
	log          *slog.Logger
	mockStore    *mocks.MockStore
	mockPlatform *mocks.MockPlatform
	service      *EpisodeService
	tempDir      string
	publicDir    string
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestEpisodeServiceSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.log = slog.Default()

	// Create temporary directories for testing
	var err error
	suite.tempDir, err = os.MkdirTemp("", "voxify_test_download_*")
	suite.Require().NoError(err)

	suite.publicDir, err = os.MkdirTemp("", "voxify_test_public_*")
	suite.Require().NoError(err)

	suite.cfg = config.Settings{
		DownloadDir:     suite.tempDir,
		PublicDir:       suite.publicDir,
		DownloadTimeout: 30 * time.Second,
		DownloadFormat:  domain.DownloadMp3,
		DownloadQuality: "192k",
	}
}

// TearDownSuite is called once after the entire test suite runs
func (suite *TestEpisodeServiceSuite) TearDownSuite() {
	if err := os.RemoveAll(suite.tempDir); err != nil {
		suite.Require().NoError(err)
	}
	if err := os.RemoveAll(suite.publicDir); err != nil {
		suite.Require().NoError(err)
	}
}

// SetupTest is called before each test method
func (suite *TestEpisodeServiceSuite) SetupSubTest() {
	suite.mockStore = mocks.NewMockStore(suite.T())
	suite.mockPlatform = mocks.NewMockPlatform(suite.T())

	suite.service = NewEpisodeService(suite.cfg, suite.log, suite.mockStore, suite.mockPlatform)
}

// TestNewEpisodeService tests the constructor
func (suite *TestEpisodeServiceSuite) TestNewEpisodeService() {
	// Act
	service := NewEpisodeService(suite.cfg, suite.log, suite.mockStore, suite.mockPlatform)

	// Assert
	suite.NotNil(service)
	suite.Equal(suite.cfg, service.cfg)
	suite.Equal(suite.mockStore, service.store)
	suite.Len(service.platforms, 1)
	suite.Equal(suite.mockPlatform, service.platforms[0])
	suite.NotNil(service.log)
}

// TestInit tests the Init method
func (suite *TestEpisodeServiceSuite) TestInit() {
	suite.Run("Success", func() {
		// Arrange
		suite.mockPlatform.On("Init", suite.ctx).Return(nil)

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.NoError(err)
	})

	suite.Run("PlatformInitFails", func() {
		// Arrange
		platformErr := errors.New("platform init error")
		suite.mockPlatform.On("Init", suite.ctx).Return(platformErr)

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "platform check failed")
		suite.Contains(err.Error(), platformErr.Error())
	})

	suite.Run("PublicDirNotExists", func() {
		// Arrange
		originalDir := suite.service.cfg.PublicDir
		suite.service.cfg.PublicDir = "/non/existent/directory"
		defer func() { suite.service.cfg.PublicDir = originalDir }()

		suite.mockPlatform.On("Init", suite.ctx).Return(nil)

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "public directory check failed")
	})

	suite.Run("DownloadDirNotExists", func() {
		// Arrange
		originalDir := suite.service.cfg.DownloadDir
		suite.service.cfg.DownloadDir = "/non/existent/directory"
		defer func() { suite.service.cfg.DownloadDir = originalDir }()

		suite.mockPlatform.On("Init", suite.ctx).Return(nil)

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "download directory check failed")
	})
}

// TestValidate tests the Validate method
func (suite *TestEpisodeServiceSuite) TestValidate() {
	req := domain.DownloadRequest{
		ID: "test-req-123",
		Source: domain.RequestSource{
			UserID:    123,
			ChatID:    456,
			MessageID: 789,
		},
		Url:             "https://youtube.com/watch?v=test123",
		DownloadFormat:  domain.DownloadMp3,
		DownloadQuality: "best",
	}

	suite.Run("Success", func() {
		// Arrange
		suite.mockPlatform.On("Match", req).Return(true)

		// Act
		err := suite.service.Validate(suite.ctx, req)

		// Assert
		suite.NoError(err)
	})

	suite.Run("NoMatchingPlatform", func() {
		// Arrange
		suite.mockPlatform.On("Match", req).Return(false)

		// Act
		err := suite.service.Validate(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Equal(domain.ErrDownloadPlatform, err)
	})

	suite.Run("InvalidRequest", func() {
		// Arrange
		invalidReq := domain.DownloadRequest{
			ID: "test-req-invalid",
			Source: domain.RequestSource{
				UserID:    123,
				ChatID:    456,
				MessageID: 789,
			},
			Url:             "invalid-url",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "best",
		}

		// Act
		err := suite.service.Validate(suite.ctx, invalidReq)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrDownloadRequest))
	})
}

// TestDownload tests the Download method
func (suite *TestEpisodeServiceSuite) TestDownload() {
	req := domain.DownloadRequest{
		ID: "test-req-123",
		Source: domain.RequestSource{
			UserID:    123,
			ChatID:    456,
			MessageID: 789,
		},
		Url:             "https://youtube.com/watch?v=test123",
		DownloadFormat:  domain.DownloadMp3,
		DownloadQuality: "best",
	}

	suite.Run("SuccessfulDownload", func() {
		// Arrange
		suite.mockPlatform.On("ID").Return("test-platform")
		suite.mockPlatform.On("Match", req).Return(true)
		suite.mockPlatform.On("Download", mock.AnythingOfType("*context.timerCtx"), req).Return(&domain.Episode{
			Title:         "Test Episode",
			Description:   "Test Description",
			MediaFile:     "audio_test.mp3",
			ThumbnailFile: "thumb_test.jpg",
			MediaType:     domain.MediaMp3,
			MediaSize:     1024000,
			MediaDuration: 3600,
			OriginalURL:   req.Url,
		}, nil)
		suite.mockStore.On("EpisodeCreate", suite.ctx, mock.MatchedBy(func(episode *domain.Episode) bool {
			return episode.OriginalURL == req.Url && episode.Title == "Test Episode"
		})).Return(nil).Run(func(args mock.Arguments) {
			// Simulate store setting ID and timestamps
			episode := args.Get(1).(*domain.Episode)
			episode.ID = 1
		})

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.NoError(err)
		suite.NotNil(result)
		suite.Equal(req.Url, result.OriginalURL)
		suite.Equal("Test Episode", result.Title)
		suite.Equal("Test Description", result.Description)
		suite.Equal("audio_test.mp3", result.MediaFile)
		suite.Equal("thumb_test.jpg", result.ThumbnailFile)
		suite.Equal(domain.MediaMp3, result.MediaType)
		suite.Equal(int64(1024000), result.MediaSize)
		suite.Equal(int64(3600), result.MediaDuration)
		suite.Equal(int64(1), result.ID)
	})

	suite.Run("NoMatchingPlatform", func() {
		// Arrange
		suite.mockPlatform.On("Match", req).Return(false)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.Equal(domain.ErrDownloadPlatform, err)
	})

	suite.Run("PlatformDownloadFails", func() {
		// Arrange
		platformErr := errors.New("platform download error")
		suite.mockPlatform.On("ID").Return("test-platform")
		suite.mockPlatform.On("Match", req).Return(true)
		suite.mockPlatform.On("Download", mock.AnythingOfType("*context.timerCtx"), req).
			Return(nil, platformErr)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.True(errors.Is(err, domain.ErrDownloadFailed))
		suite.Contains(err.Error(), platformErr.Error())
	})

	suite.Run("StoreCreateFails", func() {
		// Arrange
		storeErr := errors.New("store create error")
		suite.mockPlatform.On("ID").Return("test-platform")
		suite.mockPlatform.On("Match", req).Return(true)
		suite.mockPlatform.On("Download", mock.AnythingOfType("*context.timerCtx"), req).
			Return(&domain.Episode{
				Title:       "Test Episode",
				MediaFile:   "audio_test.mp3",
				OriginalURL: req.Url,
			}, nil)
		suite.mockStore.On("EpisodeCreate", suite.ctx, mock.AnythingOfType("*domain.Episode")).
			Return(storeErr)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.True(errors.Is(err, domain.ErrEpisodeCreate))
		suite.Contains(err.Error(), storeErr.Error())
	})

	suite.Run("ContextTimeout", func() {
		// Arrange
		shortCtx, cancel := context.WithTimeout(suite.ctx, 10*time.Millisecond)
		defer cancel()

		suite.mockPlatform.On("ID").Return("test-platform")
		suite.mockPlatform.On("Match", req).Return(true)
		suite.mockPlatform.On("Download", mock.AnythingOfType("*context.cancelCtx"), req).
			Return(nil, context.DeadlineExceeded)

		// Act
		result, err := suite.service.Download(shortCtx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.True(errors.Is(err, domain.ErrDownloadFailed))
		suite.Contains(err.Error(), context.DeadlineExceeded.Error())
	})
}

// TestFindPlatform tests the findPlatform method indirectly through Download
func (suite *TestEpisodeServiceSuite) TestFindPlatform() {
	suite.Run("MultiplePlatforms", func() {
		// Arrange
		mockPlatform2 := mocks.NewMockPlatform(suite.T())
		service := NewEpisodeService(suite.cfg, suite.log, suite.mockStore, suite.mockPlatform, mockPlatform2)

		req := domain.DownloadRequest{
			ID: "test-req-multi",
			Source: domain.RequestSource{
				UserID:    123,
				ChatID:    456,
				MessageID: 789,
			},
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "best",
		}

		// First platform doesn't match, second does
		suite.mockPlatform.On("Match", req).Return(false)
		mockPlatform2.On("ID").Return("test-platform-2")
		mockPlatform2.On("Match", req).Return(true)
		mockPlatform2.On("Download", mock.AnythingOfType("*context.timerCtx"), req).
			Return(&domain.Episode{
				Title:       "Test Episode",
				MediaFile:   "test.mp3",
				OriginalURL: req.Url,
			}, nil)
		suite.mockStore.On("EpisodeCreate", suite.ctx, mock.AnythingOfType("*domain.Episode")).
			Return(nil).Run(func(args mock.Arguments) {
			episode := args.Get(1).(*domain.Episode)
			episode.ID = 1
		})

		// Act
		result, err := service.Download(suite.ctx, req)

		// Assert
		suite.NoError(err)
		suite.NotNil(result)
		suite.Equal("Test Episode", result.Title)
		suite.Equal("test.mp3", result.MediaFile)

		// Verify that the second platform was used
		mockPlatform2.AssertExpectations(suite.T())
	})

	suite.Run("NoPlatforms", func() {
		// Arrange
		service := NewEpisodeService(suite.cfg, suite.log, suite.mockStore) // No platforms
		req := domain.DownloadRequest{
			ID: "test-req-no-platforms",
			Source: domain.RequestSource{
				UserID:    123,
				ChatID:    456,
				MessageID: 789,
			},
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "best",
		}

		// Act
		result, err := service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.Equal(domain.ErrDownloadPlatform, err)
	})
}

// TestValidateRequest tests the validateRequest method
func (suite *TestEpisodeServiceSuite) TestValidateRequest() {
	// create a service instance explicitly to ensure cfg is applied
	service := NewEpisodeService(suite.cfg, suite.log, suite.mockStore, suite.mockPlatform)

	suite.Run("Success", func() {
		req := domain.DownloadRequest{
			ID:              "req-success",
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "128k",
		}
		err := service.validateRequest(&req)
		suite.NoError(err)
	})

	suite.Run("InvalidURL", func() {
		req := domain.DownloadRequest{
			ID:              "req-bad-url",
			Url:             "ht!tp://bad-url",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "128k",
		}
		err := service.validateRequest(&req)
		suite.Error(err)
		suite.Contains(err.Error(), "url validation failed")
	})

	suite.Run("UnsupportedFormat", func() {
		req := domain.DownloadRequest{
			ID:              "req-bad-format",
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadFormat("wav"),
			DownloadQuality: "128k",
		}
		err := service.validateRequest(&req)
		suite.Error(err)
		suite.Contains(err.Error(), "download format validation failed")
	})

	suite.Run("UnsupportedQuality", func() {
		req := domain.DownloadRequest{
			ID:              "req-bad-quality",
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: ";rm -rf /",
		}
		err := service.validateRequest(&req)
		suite.Error(err)
		suite.Contains(err.Error(), "download quality validation failed")
	})

	suite.Run("EmptyFormat", func() {
		req := domain.DownloadRequest{
			ID:              "req-empty-format",
			Url:             "https://example.com/video",
			DownloadFormat:  "",
			DownloadQuality: "128k",
		}
		err := service.validateRequest(&req)
		suite.Error(err)
		suite.Contains(err.Error(), "download format not specified")
	})

	suite.Run("EmptyQuality", func() {
		req := domain.DownloadRequest{
			ID:              "req-empty-quality",
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "",
		}
		err := service.validateRequest(&req)
		suite.Error(err)
		suite.Contains(err.Error(), "download quality is empty")
	})
}
