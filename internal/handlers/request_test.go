package handlers

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/mocks"
)

// TestRequestHandlers is the entry point for running the test suite
func TestRequestHandlers(t *testing.T) {
	suite.Run(t, new(TestRequestHandlersSuite))
}

// TestRequestHandlersSuite is a test suite for RequestHandlers
type TestRequestHandlersSuite struct {
	suite.Suite
	ctx        context.Context
	cancel     context.CancelFunc
	cfg        config.Settings
	log        *slog.Logger
	bus        *mocks.MockEventBus
	builder    *mocks.MockFeedBuilder
	downloader *mocks.MockEpisodeDownloader
	handlers   *RequestHandlers
}

// SetupTest is called once before each test method
func (suite *TestRequestHandlersSuite) SetupTest() {
	suite.SetupSubTest()
}

// SetupSubTest is called before each subtest in the suite
func (suite *TestRequestHandlersSuite) SetupSubTest() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.cfg = config.Settings{
		DownloadWorkers: 2,
		DownloadDir:     "/tmp/downloads",
		PublicDir:       "/tmp/public",
	}
	suite.log = slog.New(slog.NewTextHandler(nil, &slog.HandlerOptions{Level: slog.LevelError}))
	suite.bus = mocks.NewMockEventBus(suite.T())
	suite.builder = mocks.NewMockFeedBuilder(suite.T())
	suite.downloader = mocks.NewMockEpisodeDownloader(suite.T())

	suite.handlers = NewRequestHandlers(suite.cfg, suite.log, suite.bus).
		WithBuilder(suite.builder).
		WithDownloader(suite.downloader)
}

// TearDownTest is called after each test method
func (suite *TestRequestHandlersSuite) TearDownTest() {
	suite.TearDownSubTest()
}

// TearDownSubTest is called after each subtest
func (suite *TestRequestHandlersSuite) TearDownSubTest() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// Helper functions

func (suite *TestRequestHandlersSuite) createDownloadRequest() domain.DownloadRequest {
	return domain.DownloadRequest{
		ID:              "test-id-1",
		Url:             "https://example.com/video",
		DownloadFormat:  domain.DownloadMp3,
		DownloadQuality: "192k",
		Source: domain.RequestSource{
			UserID:    12345,
			ChatID:    67890,
			MessageID: 111,
		},
	}
}

func (suite *TestRequestHandlersSuite) createBuildRequest() domain.BuildRequest {
	return domain.BuildRequest{
		ID: "build-id-1",
	}
}

func (suite *TestRequestHandlersSuite) createFeedInfoRequest() domain.FeedInfoRequest {
	return domain.FeedInfoRequest{
		ID: "info-id-1",
		Source: domain.RequestSource{
			UserID:    12345,
			ChatID:    67890,
			MessageID: 222,
		},
	}
}

func (suite *TestRequestHandlersSuite) createEpisode() *domain.Episode {
	return &domain.Episode{
		ID:            1,
		Title:         "Test Episode",
		Description:   "Test Description",
		MediaFile:     "test.mp3",
		ThumbnailFile: "test.jpg",
		MediaDuration: 180,
		MediaType:     domain.MediaMp3,
		MediaSize:     1024000,
	}
}

func (suite *TestRequestHandlersSuite) createFeedInfo() *domain.FeedInfo {
	return &domain.FeedInfo{
		Title:       "Test Feed",
		Description: "Test Feed Description",
		RSSLink:     "https://example.com/feed",
		WebsiteLink: "https://example.com",
	}
}

// TestNewRequestHandlers tests the constructor
func (suite *TestRequestHandlersSuite) TestNewRequestHandlers() {
	suite.Run("Success", func() {
		// Act
		h := NewRequestHandlers(suite.cfg, suite.log, suite.bus)

		// Assert
		suite.NotNil(h)
		suite.Equal(suite.cfg, h.cfg)
		suite.Equal(suite.log, h.log)
		suite.Equal(suite.bus, h.bus)
		suite.NotNil(h.queue)
		suite.Nil(h.builder)
		suite.Nil(h.downloader)
	})
}

// TestWithBuilder tests the WithBuilder method
func (suite *TestRequestHandlersSuite) TestWithBuilder() {
	suite.Run("Success", func() {
		// Arrange
		h := NewRequestHandlers(suite.cfg, suite.log, suite.bus)

		// Act
		result := h.WithBuilder(suite.builder)

		// Assert
		suite.Equal(h, result)
		suite.Equal(suite.builder, h.builder)
	})
}

// TestWithDownloader tests the WithDownloader method
func (suite *TestRequestHandlersSuite) TestWithDownloader() {
	suite.Run("Success", func() {
		// Arrange
		h := NewRequestHandlers(suite.cfg, suite.log, suite.bus)

		// Act
		result := h.WithDownloader(suite.downloader)

		// Assert
		suite.Equal(h, result)
		suite.Equal(suite.downloader, h.downloader)
	})
}

// TestStart tests the Start method
func (suite *TestRequestHandlersSuite) TestStart() {
	suite.Run("Success", func() {
		// Arrange
		suite.bus.On("Subscribe", domain.DownloadRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.On("Subscribe", domain.BuildRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.On("Subscribe", domain.FeedInfoRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.On("Subscribe", domain.DownloadResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.On("Subscribe", domain.BuildResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.On("Subscribe", domain.FeedInfoResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()

		// Act
		err := suite.handlers.Init(suite.ctx)

		// Assert
		suite.NoError(err)
		suite.bus.AssertExpectations(suite.T())

		// Cleanup
		suite.cancel()
		suite.handlers.Wait()
	})

	suite.Run("ErrorWhenBuilderNotSet", func() {
		// Arrange
		h := NewRequestHandlers(suite.cfg, suite.log, suite.bus).WithDownloader(suite.downloader)

		// Act
		err := h.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "feed builder is not set")
	})

	suite.Run("ErrorWhenDownloaderNotSet", func() {
		// Arrange
		h := NewRequestHandlers(suite.cfg, suite.log, suite.bus).WithBuilder(suite.builder)

		// Act
		err := h.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "episode downloader is not set")
	})

	suite.Run("ErrorWhenBusNotSet", func() {
		// Arrange
		h := NewRequestHandlers(suite.cfg, suite.log, nil).
			WithBuilder(suite.builder).
			WithDownloader(suite.downloader)

		// Act
		err := h.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "event bus is not set")
	})
}

// TestDownloadHandler tests the download request handler
func (suite *TestRequestHandlersSuite) TestDownloadHandler() {
	suite.Run("SuccessEnqueue", func() {
		// Arrange
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)

		suite.downloader.On("Validate", suite.ctx, req).Return(nil)
		// We expect Publish to be called when request is successfully enqueued
		suite.bus.On("Publish", mock.Anything).Return().Once()

		handler := suite.handlers.downloadHandler(suite.ctx)

		// Start a goroutine to consume from the queue (simulating a worker)
		receivedReq := make(chan domain.DownloadRequest, 1)
		consumerReady := make(chan struct{})
		producerSent := make(chan struct{})
		go func() {
			close(consumerReady) // Signal that consumer is ready
			select {
			case r := <-suite.handlers.queue:
				receivedReq <- r
			case <-time.After(time.Second):
				// Timeout - nothing was sent to queue
			}
		}()

		// Act - call handler in goroutine to avoid blocking on unbuffered channel
		go func() {
			<-consumerReady                    // Wait for consumer to be ready
			time.Sleep(200 * time.Millisecond) // Ensure consumer is waiting
			handler(event)
			close(producerSent) // Signal that producer has sent the request
		}()

		<-producerSent // Wait for producer to finish

		// Assert - verify the request was successfully enqueued
		select {
		case r := <-receivedReq:
			suite.Equal(req.ID, r.ID, "Request should be enqueued with correct ID")
			suite.Equal(req.Url, r.Url, "Request should be enqueued with correct URL")
		case <-time.After(time.Second):
			suite.Fail("Request was not enqueued - timeout waiting for request in queue")
		}

		// Verify the mocks were called
		suite.downloader.AssertExpectations(suite.T())
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("ValidationError", func() {
		// Arrange
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)
		validationErr := errors.New("validation failed")

		suite.downloader.On("Validate", suite.ctx, req).Return(validationErr)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusFailed && resp.Error != nil
		})).Return()

		handler := suite.handlers.downloadHandler(suite.ctx)

		// Act
		handler(event)

		// Assert - should not be enqueued
		suite.Len(suite.handlers.queue, 0)
	})

	suite.Run("AlreadyInProgress", func() {
		// Arrange
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)

		suite.handlers.active.Store(req.Url, struct{}{})

		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, domain.ErrDownloadInProgress)
		})).Return()

		handler := suite.handlers.downloadHandler(suite.ctx)

		// Act
		handler(event)

		// Assert
		suite.Len(suite.handlers.queue, 0)
	})

	suite.Run("QueueFull", func() {
		// Arrange
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)

		// Create a fresh handler with small queue
		smallCfg := suite.cfg
		smallCfg.DownloadWorkers = 1
		h := NewRequestHandlers(smallCfg, suite.log, suite.bus).
			WithBuilder(suite.builder).
			WithDownloader(suite.downloader)

		// Fill the queue without blocking
		dummyReq := suite.createDownloadRequest()
		dummyReq.Url = "https://example.com/other"
		select {
		case h.queue <- dummyReq:
		default:
		}

		// Now the queue should be full
		suite.downloader.On("Validate", suite.ctx, req).Return(nil)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, domain.ErrDownloadBusy)
		})).Return()

		handler := h.downloadHandler(suite.ctx)

		// Act
		handler(event)

		// Assert - the request should fail with ErrDownloadBusy
		suite.bus.AssertExpectations(suite.T())
	})
}

// TestDownload tests the download method
func (suite *TestRequestHandlersSuite) TestDownload() {
	suite.Run("Success", func() {
		// Arrange
		req := suite.createDownloadRequest()
		episode := suite.createEpisode()

		suite.downloader.On("Download", suite.ctx, req).Return(episode, nil)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusSuccess && resp.Episode == episode
		})).Return()
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.BuildRequestEvent {
				return false
			}
			buildReq := e.Payload().(domain.BuildRequest)
			return buildReq.ID == req.ID
		})).Return()

		// Act
		suite.handlers.download(suite.ctx, req)

		// Assert
		_, exists := suite.handlers.active.Load(req.Url)
		suite.False(exists, "URL should be removed from active downloads")
	})

	suite.Run("AlreadyDownloading", func() {
		// Arrange
		req := suite.createDownloadRequest()
		suite.handlers.active.Store(req.Url, struct{}{})

		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, domain.ErrDownloadInProgress)
		})).Return()

		// Act
		suite.handlers.download(suite.ctx, req)

		// Assert - should still be in active
		_, exists := suite.handlers.active.Load(req.Url)
		suite.True(exists)
	})

	suite.Run("DownloadError", func() {
		// Arrange
		req := suite.createDownloadRequest()
		downloadErr := errors.New("download failed")

		suite.downloader.On("Download", suite.ctx, req).Return(nil, downloadErr)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusFailed && resp.Error != nil
		})).Return()

		// Act
		suite.handlers.download(suite.ctx, req)

		// Assert
		_, exists := suite.handlers.active.Load(req.Url)
		suite.False(exists, "URL should be removed from active downloads even on error")
	})

	suite.Run("ContextCancelled", func() {
		// Arrange
		req := suite.createDownloadRequest()
		ctx, cancel := context.WithCancel(suite.ctx)
		cancel()

		downloadErr := errors.New("operation cancelled")

		suite.downloader.On("Download", ctx, req).Return(nil, downloadErr)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, domain.ErrDownloadInterrupted)
		})).Return()

		// Act
		suite.handlers.download(ctx, req)

		// Assert
		_, exists := suite.handlers.active.Load(req.Url)
		suite.False(exists)
	})
}

// TestBuildHandler tests the build request handler
func (suite *TestRequestHandlersSuite) TestBuildHandler() {
	suite.Run("Success", func() {
		// Arrange
		req := suite.createBuildRequest()
		event := domain.NewBuildRequestEvent(req)

		suite.builder.On("Build", suite.ctx).Return(nil)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.BuildResponseEvent {
				return false
			}
			resp := e.Payload().(domain.BuildResponse)
			return resp.Status == domain.StatusSuccess && resp.Request.ID == req.ID
		})).Return()

		handler := suite.handlers.buildHandler(suite.ctx)

		// Act
		handler(event)

		// Assert
		suite.builder.AssertExpectations(suite.T())
	})

	suite.Run("BuildError", func() {
		// Arrange
		req := suite.createBuildRequest()
		event := domain.NewBuildRequestEvent(req)
		buildErr := errors.New("build failed")

		suite.builder.On("Build", suite.ctx).Return(buildErr)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.BuildResponseEvent {
				return false
			}
			resp := e.Payload().(domain.BuildResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, buildErr)
		})).Return()

		handler := suite.handlers.buildHandler(suite.ctx)

		// Act
		handler(event)

		// Assert
		suite.builder.AssertExpectations(suite.T())
	})
}

// TestFeedInfoHandler tests the feed info request handler
func (suite *TestRequestHandlersSuite) TestFeedInfoHandler() {
	suite.Run("Success", func() {
		// Arrange
		req := suite.createFeedInfoRequest()
		event := domain.NewFeedInfoRequestEvent(req)
		feedInfo := suite.createFeedInfo()

		suite.builder.On("Info", suite.ctx).Return(feedInfo, nil)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.FeedInfoResponseEvent {
				return false
			}
			resp := e.Payload().(domain.FeedInfoResponse)
			return resp.Status == domain.StatusSuccess && resp.FeedInfo == feedInfo
		})).Return()

		handler := suite.handlers.feedInfoHandler(suite.ctx)

		// Act
		handler(event)

		// Assert
		suite.builder.AssertExpectations(suite.T())
	})

	suite.Run("InfoError", func() {
		// Arrange
		req := suite.createFeedInfoRequest()
		event := domain.NewFeedInfoRequestEvent(req)
		infoErr := errors.New("info failed")

		suite.builder.On("Info", suite.ctx).Return(nil, infoErr)
		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.FeedInfoResponseEvent {
				return false
			}
			resp := e.Payload().(domain.FeedInfoResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, infoErr)
		})).Return()

		handler := suite.handlers.feedInfoHandler(suite.ctx)

		// Act
		handler(event)

		// Assert
		suite.builder.AssertExpectations(suite.T())
	})
}

// TestDownloadWorker tests the download worker
func (suite *TestRequestHandlersSuite) TestDownloadWorker() {
	suite.Run("ProcessesRequestsViaDownloadMethod", func() {
		// Arrange
		req := suite.createDownloadRequest()
		episode := suite.createEpisode()

		suite.downloader.On("Download", suite.ctx, req).Return(episode, nil)
		suite.bus.On("Publish", mock.Anything).Return()

		// Act - directly test download method (which is what worker calls)
		suite.handlers.download(suite.ctx, req)

		// Assert
		suite.downloader.AssertExpectations(suite.T())
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("StopsOnContextCancel", func() {
		// Arrange
		suite.bus.On("Subscribe", mock.Anything, mock.Anything).Return()

		err := suite.handlers.Init(suite.ctx)
		suite.Require().NoError(err)

		// Act - cancel context
		suite.cancel()

		// Assert - Wait should complete without hanging
		done := make(chan struct{})
		go func() {
			suite.handlers.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - workers stopped gracefully
		case <-time.After(time.Second):
			suite.Fail("Workers did not stop in time")
		}
	})
}

// TestFailRequest tests the failRequest method
func (suite *TestRequestHandlersSuite) TestFailRequest() {
	suite.Run("DownloadRequest", func() {
		// Arrange
		req := suite.createDownloadRequest()
		testErr := errors.New("test error")

		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp := e.Payload().(domain.DownloadResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, testErr)
		})).Return()

		// Act
		suite.handlers.failRequest(req, testErr)

		// Assert
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("BuildRequest", func() {
		// Arrange
		req := suite.createBuildRequest()
		testErr := errors.New("test error")

		suite.bus.On("Publish", mock.MatchedBy(func(e domain.Event) bool {
			if e.Type() != domain.BuildResponseEvent {
				return false
			}
			resp := e.Payload().(domain.BuildResponse)
			return resp.Status == domain.StatusFailed && errors.Is(resp.Error, testErr)
		})).Return()

		// Act
		suite.handlers.failRequest(req, testErr)

		// Assert
		suite.bus.AssertExpectations(suite.T())
	})
}

// TestLoggingHandlers tests logging handlers
func (suite *TestRequestHandlersSuite) TestLoggingHandlers() {
	suite.Run("LogDownloadRequestHandler", func() {
		// Arrange
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)

		// Act & Assert - should not panic
		suite.NotPanics(func() {
			suite.handlers.logDownloadRequestHandler(event)
		})
	})

	suite.Run("LogDownloadResponseHandler", func() {
		// Arrange
		resp := domain.DownloadResponse{
			Status:  domain.StatusSuccess,
			Request: suite.createDownloadRequest(),
			Episode: suite.createEpisode(),
		}
		event := domain.NewDownloadResponseEvent(resp)

		// Act & Assert - should not panic
		suite.NotPanics(func() {
			suite.handlers.logDownloadResponseHandler(event)
		})
	})

	suite.Run("LogBuildRequestHandler", func() {
		// Arrange
		req := suite.createBuildRequest()
		event := domain.NewBuildRequestEvent(req)

		// Act & Assert - should not panic
		suite.NotPanics(func() {
			suite.handlers.logBuildRequestHandler(event)
		})
	})

	suite.Run("LogBuildResponseHandler", func() {
		// Arrange
		resp := domain.BuildResponse{
			Status:  domain.StatusSuccess,
			Request: suite.createBuildRequest(),
		}
		event := domain.NewBuildResponseEvent(resp)

		// Act & Assert - should not panic
		suite.NotPanics(func() {
			suite.handlers.logBuildResponseHandler(event)
		})
	})

	suite.Run("LogFeedInfoRequestHandler", func() {
		// Arrange
		req := suite.createFeedInfoRequest()
		event := domain.NewFeedInfoRequestEvent(req)

		// Act & Assert - should not panic
		suite.NotPanics(func() {
			suite.handlers.logFeedInfoRequestHandler(event)
		})
	})

	suite.Run("LogFeedInfoResponseHandler", func() {
		// Arrange
		resp := domain.FeedInfoResponse{
			Status:   domain.StatusSuccess,
			Request:  suite.createFeedInfoRequest(),
			FeedInfo: suite.createFeedInfo(),
		}
		event := domain.NewFeedInfoResponseEvent(resp)

		// Act & Assert - should not panic
		suite.NotPanics(func() {
			suite.handlers.logFeedInfoResponseHandler(event)
		})
	})
}

// TestWait tests the Wait method
func (suite *TestRequestHandlersSuite) TestWait() {
	suite.Run("WaitsForWorkers", func() {
		// Arrange
		suite.bus.On("Subscribe", mock.Anything, mock.Anything).Return()

		err := suite.handlers.Init(suite.ctx)
		suite.Require().NoError(err)

		// Act
		done := make(chan struct{})
		go func() {
			suite.handlers.Wait()
			close(done)
		}()

		// Cancel context to stop workers
		suite.cancel()

		// Assert - Wait should complete
		select {
		case <-done:
			// Success
		case <-time.After(2 * time.Second):
			suite.Fail("Wait() did not complete in time")
		}
	})
}
