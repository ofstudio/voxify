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

// SetupSubTest is called before each subtest in the suite
func (suite *TestEpisodeServiceSuite) SetupSubTest() {
	suite.Require().NoError(os.RemoveAll(suite.tempDir))
	suite.Require().NoError(os.MkdirAll(suite.tempDir, 0755))
	suite.Require().NoError(os.RemoveAll(suite.publicDir))
	suite.Require().NoError(os.MkdirAll(suite.publicDir, 0755))

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
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(0, nil)

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.NoError(err)
	})

	suite.Run("FeedLimitDeletesOldEpisodes", func() {
		// Arrange
		suite.service.cfg.FeedMaxEpisodes = 1
		oldEpisode := &domain.Episode{
			ID:            1,
			MediaFile:     "old.mp3",
			ThumbnailFile: "old.jpg",
			OriginalURL:   "https://example.com/old",
		}
		suite.Require().NoError(os.WriteFile(suite.publicFile(oldEpisode.MediaFile), []byte("media"), 0644))
		suite.Require().NoError(os.WriteFile(suite.publicFile(oldEpisode.ThumbnailFile), []byte("thumb"), 0644))

		suite.mockPlatform.On("Init", suite.ctx).Return(nil)
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(2, nil)
		suite.mockStore.On("EpisodeGetOldest", suite.ctx, 1).Return([]*domain.Episode{oldEpisode}, nil)
		suite.mockStore.On("EpisodeDelete", suite.ctx, oldEpisode.ID).Return(nil)

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.NoError(err)
		suite.NoFileExists(suite.publicFile(oldEpisode.MediaFile))
		suite.NoFileExists(suite.publicFile(oldEpisode.ThumbnailFile))
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
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
		suite.mockPlatform.On("Match", req).Return(true)

		// Act
		err := suite.service.Validate(suite.ctx, req)

		// Assert
		suite.NoError(err)
	})

	suite.Run("NoMatchingPlatform", func() {
		// Arrange
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
		suite.mockPlatform.On("Match", req).Return(false)

		// Act
		err := suite.service.Validate(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Equal(domain.ErrNoMatchingPlatform, err)
	})

	suite.Run("InvalidUrl", func() {
		// Arrange
		invalidReq := domain.DownloadRequest{
			ID:              "test-req-invalid",
			Source:          domain.RequestSource{UserID: 123, ChatID: 456, MessageID: 789},
			Url:             "invalid-url",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "best",
		}
		// removed store expectation because validateRequest returns earlier on invalid URL

		// Act
		err := suite.service.Validate(suite.ctx, invalidReq)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrInvalidUrl))
	})

	suite.Run("DuplicateExists", func() {
		// Arrange
		existing := []*domain.Episode{{ID: 1, OriginalURL: req.Url, Title: "Existing"}}
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return(existing, nil)

		// Act
		err := suite.service.Validate(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrDownloadExists))
	})

	suite.Run("StoreError", func() {
		// Arrange
		storeErr := errors.New("db failure")
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return(nil, storeErr)

		// Act
		err := suite.service.Validate(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, storeErr))
		suite.Contains(err.Error(), storeErr.Error())
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
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
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
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(1, nil)

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

	suite.Run("SuccessfulDownloadDeletesOldEpisodes", func() {
		// Arrange
		suite.service.cfg.FeedMaxEpisodes = 1
		oldEpisode := &domain.Episode{
			ID:            10,
			MediaFile:     "download-old.mp3",
			ThumbnailFile: "download-old.jpg",
			OriginalURL:   "https://example.com/old",
		}
		suite.Require().NoError(os.WriteFile(suite.publicFile(oldEpisode.MediaFile), []byte("media"), 0644))
		suite.Require().NoError(os.WriteFile(suite.publicFile(oldEpisode.ThumbnailFile), []byte("thumb"), 0644))

		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
		suite.mockPlatform.On("ID").Return("test-platform")
		suite.mockPlatform.On("Match", req).Return(true)
		suite.mockPlatform.On("Download", mock.AnythingOfType("*context.timerCtx"), req).Return(&domain.Episode{
			Title:       "New Episode",
			MediaFile:   "new.mp3",
			MediaType:   domain.MediaMp3,
			OriginalURL: req.Url,
		}, nil)
		suite.mockStore.On("EpisodeCreate", suite.ctx, mock.AnythingOfType("*domain.Episode")).
			Return(nil).Run(func(args mock.Arguments) {
			episode := args.Get(1).(*domain.Episode)
			episode.ID = 11
		})
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(2, nil)
		suite.mockStore.On("EpisodeGetOldest", suite.ctx, 1).Return([]*domain.Episode{oldEpisode}, nil)
		suite.mockStore.On("EpisodeDelete", suite.ctx, oldEpisode.ID).Return(nil)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.NoError(err)
		suite.NotNil(result)
		suite.Equal(int64(11), result.ID)
		suite.NoFileExists(suite.publicFile(oldEpisode.MediaFile))
		suite.NoFileExists(suite.publicFile(oldEpisode.ThumbnailFile))
	})

	suite.Run("NoMatchingPlatform", func() {
		// Arrange
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
		suite.mockPlatform.On("Match", req).Return(false)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.Equal(domain.ErrNoMatchingPlatform, err)
	})

	suite.Run("PlatformDownloadFails", func() {
		// Arrange
		platformErr := errors.New("platform download error")
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
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
		episode := &domain.Episode{
			Title:         "Test Episode",
			MediaFile:     "store-fail.mp3",
			ThumbnailFile: "store-fail.jpg",
			OriginalURL:   req.Url,
		}
		suite.Require().NoError(os.WriteFile(suite.publicFile(episode.MediaFile), []byte("media"), 0644))
		suite.Require().NoError(os.WriteFile(suite.publicFile(episode.ThumbnailFile), []byte("thumb"), 0644))

		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
		suite.mockPlatform.On("ID").Return("test-platform")
		suite.mockPlatform.On("Match", req).Return(true)
		suite.mockPlatform.On("Download", mock.AnythingOfType("*context.timerCtx"), req).
			Return(episode, nil)
		suite.mockStore.On("EpisodeCreate", suite.ctx, mock.AnythingOfType("*domain.Episode")).
			Return(storeErr)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.True(errors.Is(err, storeErr))
		suite.Contains(err.Error(), storeErr.Error())
		suite.NoFileExists(suite.publicFile(episode.MediaFile))
		suite.NoFileExists(suite.publicFile(episode.ThumbnailFile))
	})

	suite.Run("ContextTimeout", func() {
		// Arrange
		shortCtx, cancel := context.WithTimeout(suite.ctx, 10*time.Millisecond)
		defer cancel()

		suite.mockStore.On("EpisodeGetByUrl", shortCtx, req.Url).Return([]*domain.Episode{}, nil)
		suite.mockPlatform.On("ID").Return("test-platform")
		suite.mockPlatform.On("Match", req).Return(true)
		suite.mockPlatform.On("Download", mock.Anything, req).Return(nil, context.DeadlineExceeded)

		// Act
		result, err := suite.service.Download(shortCtx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.True(errors.Is(err, domain.ErrDownloadFailed))
		suite.Contains(err.Error(), context.DeadlineExceeded.Error())
	})

	suite.Run("DuplicateExists", func() {
		// Arrange - existing episode means validation fails early
		existing := []*domain.Episode{{ID: 1, OriginalURL: req.Url, Title: "Existing"}}
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return(existing, nil)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.True(errors.Is(err, domain.ErrDownloadExists))
	})

	suite.Run("StoreGetByUrlError", func() {
		// Arrange
		storeErr := errors.New("db failure")
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return(nil, storeErr)

		// Act
		result, err := suite.service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.True(errors.Is(err, storeErr))
		suite.Contains(err.Error(), storeErr.Error())
	})
}

// TestFindPlatform tests the findPlatform method indirectly through Download
func (suite *TestEpisodeServiceSuite) TestFindPlatform() {
	suite.Run("MultiplePlatforms", func() {
		// Arrange
		mockPlatform2 := mocks.NewMockPlatform(suite.T())
		service := NewEpisodeService(suite.cfg, suite.log, suite.mockStore, suite.mockPlatform, mockPlatform2)

		req := domain.DownloadRequest{
			ID:              "test-req-multi",
			Source:          domain.RequestSource{UserID: 123, ChatID: 456, MessageID: 789},
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "best",
		}

		// First platform doesn't match, second does
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
		suite.mockPlatform.On("Match", req).Return(false)
		mockPlatform2.On("ID").Return("test-platform-2")
		mockPlatform2.On("Match", req).Return(true)
		mockPlatform2.On("Download", mock.AnythingOfType("*context.timerCtx"), req).
			Return(&domain.Episode{Title: "Test Episode", MediaFile: "test.mp3", OriginalURL: req.Url}, nil)
		suite.mockStore.On("EpisodeCreate", suite.ctx, mock.AnythingOfType("*domain.Episode")).
			Return(nil).Run(func(args mock.Arguments) {
			episode := args.Get(1).(*domain.Episode)
			episode.ID = 1
		})
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(1, nil)

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
			ID:              "test-req-no-platforms",
			Source:          domain.RequestSource{UserID: 123, ChatID: 456, MessageID: 789},
			Url:             "https://example.com/video",
			DownloadFormat:  domain.DownloadMp3,
			DownloadQuality: "best",
		}
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)

		// Act
		result, err := service.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(result)
		suite.Equal(domain.ErrNoMatchingPlatform, err)
	})
}

func (suite *TestEpisodeServiceSuite) TestEnforceFeedMaxEpisodes() {
	suite.Run("NoLimit", func() {
		// Arrange
		suite.service.cfg.FeedMaxEpisodes = 0
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(3, nil)

		// Act
		err := suite.service.enforceFeedMaxEpisodes(suite.ctx)

		// Assert
		suite.NoError(err)
	})

	suite.Run("StoreCountError", func() {
		// Arrange
		expectedErr := errors.New("count failed")
		suite.service.cfg.FeedMaxEpisodes = 1
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(0, expectedErr)

		// Act
		err := suite.service.enforceFeedMaxEpisodes(suite.ctx)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, expectedErr))
	})

	suite.Run("DeleteFileErrorSkipsStoreDelete", func() {
		// Arrange
		suite.service.cfg.FeedMaxEpisodes = 1
		oldEpisode := &domain.Episode{ID: 1, MediaFile: "../unsafe.mp3"}
		suite.mockStore.On("EpisodeCount", suite.ctx).Return(2, nil)
		suite.mockStore.On("EpisodeGetOldest", suite.ctx, 1).Return([]*domain.Episode{oldEpisode}, nil)

		// Act
		err := suite.service.enforceFeedMaxEpisodes(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "unsafe public file path")
	})

	suite.Run("StoreDeleteErrorAfterFilesDeleted", func() {
		// Arrange
		storeErr := errors.New("delete failed")
		suite.service.cfg.FeedMaxEpisodes = 1
		oldEpisode := &domain.Episode{
			ID:            1,
			MediaFile:     "delete-store-error.mp3",
			ThumbnailFile: "delete-store-error.jpg",
		}
		suite.Require().NoError(os.WriteFile(suite.publicFile(oldEpisode.MediaFile), []byte("media"), 0644))
		suite.Require().NoError(os.WriteFile(suite.publicFile(oldEpisode.ThumbnailFile), []byte("thumb"), 0644))

		suite.mockStore.On("EpisodeCount", suite.ctx).Return(2, nil)
		suite.mockStore.On("EpisodeGetOldest", suite.ctx, 1).Return([]*domain.Episode{oldEpisode}, nil)
		suite.mockStore.On("EpisodeDelete", suite.ctx, oldEpisode.ID).Return(storeErr).Run(func(args mock.Arguments) {
			suite.NoFileExists(suite.publicFile(oldEpisode.MediaFile))
			suite.NoFileExists(suite.publicFile(oldEpisode.ThumbnailFile))
		})

		// Act
		err := suite.service.enforceFeedMaxEpisodes(suite.ctx)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, storeErr))
		suite.NoFileExists(suite.publicFile(oldEpisode.MediaFile))
		suite.NoFileExists(suite.publicFile(oldEpisode.ThumbnailFile))
	})
}

func (suite *TestEpisodeServiceSuite) publicFile(filename string) string {
	return suite.publicDir + string(os.PathSeparator) + filename
}

// TestValidateRequest tests the validateRequest method
func (suite *TestEpisodeServiceSuite) TestValidateRequest() {

	suite.Run("Success", func() {
		req := domain.DownloadRequest{ID: "req-success", Url: "https://example.com/video", DownloadFormat: domain.DownloadMp3, DownloadQuality: "128k"}
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return([]*domain.Episode{}, nil)
		err := suite.service.validateRequest(suite.ctx, req)
		suite.NoError(err)
	})

	suite.Run("InvalidURL", func() {
		req := domain.DownloadRequest{ID: "req-bad-url", Url: "ht!tp://bad-url", DownloadFormat: domain.DownloadMp3, DownloadQuality: "128k"}
		// no EpisodeGetByUrl expectation (URL validation fails first)
		err := suite.service.validateRequest(suite.ctx, req)
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrInvalidUrl))
	})

	suite.Run("UnsupportedFormat", func() {
		req := domain.DownloadRequest{ID: "req-bad-format", Url: "https://example.com/video", DownloadFormat: domain.DownloadFormat("wav"), DownloadQuality: "128k"}
		// no EpisodeGetByUrl expectation (format validation fails before store call)
		err := suite.service.validateRequest(suite.ctx, req)
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrInvalidFormat))
	})

	suite.Run("UnsupportedQuality", func() {
		req := domain.DownloadRequest{ID: "req-bad-quality", Url: "https://example.com/video", DownloadFormat: domain.DownloadMp3, DownloadQuality: ";rm -rf /"}
		// no EpisodeGetByUrl expectation (quality validation fails before store call)
		err := suite.service.validateRequest(suite.ctx, req)
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrDownloadQuality))
	})

	suite.Run("EmptyFormat", func() {
		req := domain.DownloadRequest{ID: "req-empty-format", Url: "https://example.com/video", DownloadFormat: "", DownloadQuality: "128k"}
		// no EpisodeGetByUrl expectation
		err := suite.service.validateRequest(suite.ctx, req)
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrInvalidFormat))
		suite.Contains(err.Error(), "download format not specified")
	})

	suite.Run("EmptyQuality", func() {
		req := domain.DownloadRequest{ID: "req-empty-quality", Url: "https://example.com/video", DownloadFormat: domain.DownloadMp3, DownloadQuality: ""}
		// no EpisodeGetByUrl expectation
		err := suite.service.validateRequest(suite.ctx, req)
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrDownloadQuality))
		suite.Contains(err.Error(), "download quality is empty")
	})

	suite.Run("DuplicateExists", func() {
		req := domain.DownloadRequest{ID: "req-dup", Url: "https://example.com/video", DownloadFormat: domain.DownloadMp3, DownloadQuality: "128k"}
		existing := []*domain.Episode{{ID: 5, OriginalURL: req.Url}}
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return(existing, nil)
		err := suite.service.validateRequest(suite.ctx, req)
		suite.Error(err)
		suite.True(errors.Is(err, domain.ErrDownloadExists))
	})

	suite.Run("StoreError", func() {
		req := domain.DownloadRequest{ID: "req-store-err", Url: "https://example.com/video", DownloadFormat: domain.DownloadMp3, DownloadQuality: "128k"}
		storeErr := errors.New("db failure")
		suite.mockStore.On("EpisodeGetByUrl", suite.ctx, req.Url).Return(nil, storeErr)
		err := suite.service.validateRequest(suite.ctx, req)
		suite.Error(err)
		suite.True(errors.Is(err, storeErr))
		suite.Contains(err.Error(), storeErr.Error())
	})
}
