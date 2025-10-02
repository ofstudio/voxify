package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/mocks"
)

// TestEventHandlers is the entry point for running the test suite
func TestEventHandlers(t *testing.T) {
	suite.Run(t, new(TestEventHandlersSuite))
}

// TestEventHandlersSuite is a test suite for EventHandlers
type TestEventHandlersSuite struct {
	suite.Suite
	ctx        context.Context
	cancel     context.CancelFunc
	cfg        config.Settings
	log        *slog.Logger
	bus        *mocks.MockEventBus
	builder    *mocks.MockFeedBuilder
	downloader *mocks.MockEpisodeDownloader
	handlers   *EventHandlers
}

// SetupTest is called once before each test start
func (suite *TestEventHandlersSuite) SetupTest() {
	suite.SetupSubTest()
}

// SetupSubTest is called before each subtest in the suite
func (suite *TestEventHandlersSuite) SetupSubTest() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.cfg = config.Settings{
		DownloadWorkers: 2,
	}
	suite.log = slog.Default()
	suite.bus = mocks.NewMockEventBus(suite.T())
	suite.builder = mocks.NewMockFeedBuilder(suite.T())
	suite.downloader = mocks.NewMockEpisodeDownloader(suite.T())

	suite.handlers = NewEventHandlers(
		suite.cfg,
		suite.log,
		suite.bus,
		suite.builder,
		suite.downloader,
	)
}

// TearDownTest is called after each test in the suite completes
func (suite *TestEventHandlersSuite) TearDownTest() {
	suite.TearDownSubTest()
}

// TearDownSubTest is called after each subtest in the suite
func (suite *TestEventHandlersSuite) TearDownSubTest() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// Helper functions

func (suite *TestEventHandlersSuite) createDownloadRequest() domain.DownloadRequest {
	return domain.DownloadRequest{
		ID: "test-req-1",
		Source: domain.RequestSource{
			UserID:    12345,
			ChatID:    67890,
			MessageID: 111,
		},
		Url:             "https://example.com/video",
		DownloadFormat:  domain.DownloadMp3,
		DownloadQuality: "192",
	}
}

func (suite *TestEventHandlersSuite) createBuildRequest() domain.BuildRequest {
	return domain.BuildRequest{
		ID: "build-req-1",
		Source: &domain.RequestSource{
			UserID:    12345,
			ChatID:    67890,
			MessageID: 222,
		},
	}
}

func (suite *TestEventHandlersSuite) createEpisode() *domain.Episode {
	return &domain.Episode{
		ID:            1,
		Title:         "Test Episode",
		Description:   "Test Description",
		MediaFile:     "test.mp3",
		MediaDuration: 3600,
		MediaSize:     1024000,
		MediaType:     "audio/mpeg",
		Author:        "Test Author",
		OriginalURL:   "https://example.com/video",
		CanonicalURL:  "https://example.com/canonical",
		CreatedAt:     time.Now(),
	}
}

// Test NewEventHandlers

func (suite *TestEventHandlersSuite) TestNewEventHandlers() {
	// Act
	handlers := NewEventHandlers(
		suite.cfg,
		suite.log,
		suite.bus,
		suite.builder,
		suite.downloader,
	)

	// Assert
	suite.NotNil(handlers)
	suite.Equal(suite.cfg, handlers.cfg)
	suite.Equal(suite.log, handlers.log)
	suite.Equal(suite.bus, handlers.bus)
	suite.Equal(suite.builder, handlers.builder)
	suite.Equal(suite.downloader, handlers.downloader)
	suite.NotNil(handlers.queue)
}

// Test Start method

func (suite *TestEventHandlersSuite) TestStart() {
	suite.Run("SubscribesAllHandlers", func() {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Expect all event subscriptions with correct EventType constants (cast to domain.EventType)
		suite.bus.EXPECT().Subscribe(domain.DownloadRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.BuildRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.FeedInfoRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.DownloadRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.DownloadResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.BuildRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.BuildResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.FeedInfoRequestEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.FeedInfoResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()

		// Act
		suite.handlers.Start(ctx)

		// Give some time for workers to start
		time.Sleep(10 * time.Millisecond)

		// Assert
		suite.bus.AssertExpectations(suite.T())

		// Verify workers are running by checking if they can process requests
		req := suite.createDownloadRequest()
		episode := suite.createEpisode()

		suite.downloader.EXPECT().Download(ctx, req).Return(episode, nil)
		suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusSuccess, episode, nil)).Return()

		// Send request to queue to verify workers are active
		suite.handlers.queue <- req

		// Give time for processing
		time.Sleep(50 * time.Millisecond)

		// Verify URL is not in active map after processing (worker completed)
		_, exists := suite.handlers.active.Load(req.Url)
		suite.False(exists)
	})

	suite.Run("StartsCorrectNumberOfWorkers", func() {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Set specific number of workers
		suite.cfg.DownloadWorkers = 3
		handlers := NewEventHandlers(
			suite.cfg,
			suite.log,
			suite.bus,
			suite.builder,
			suite.downloader,
		)

		// Mock all subscriptions
		suite.bus.EXPECT().Subscribe(mock.Anything, mock.Anything).Return().Times(9)

		// Act
		handlers.Start(ctx)

		// Give time for workers to start
		time.Sleep(10 * time.Millisecond)

		// Assert - verify we can process multiple requests concurrently
		// This indirectly verifies that multiple workers are running
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				req := domain.DownloadRequest{
					ID: fmt.Sprintf("test-req-%d", id),
					Source: domain.RequestSource{
						UserID:    int64(12345 + id),
						ChatID:    67890,
						MessageID: 111 + id,
					},
					Url:             fmt.Sprintf("https://example.com/video%d", id),
					DownloadFormat:  domain.DownloadMp3,
					DownloadQuality: "192",
				}
				episode := &domain.Episode{
					ID:          int64(id + 1),
					Title:       fmt.Sprintf("Test Episode %d", id),
					OriginalURL: req.Url,
				}

				suite.downloader.EXPECT().Download(ctx, req).Return(episode, nil)
				suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusSuccess, episode, nil)).Return()

				handlers.queue <- req
			}(i)
		}

		// Wait for all requests to be processed
		wg.Wait()
		time.Sleep(50 * time.Millisecond)

		// Verify all URLs are processed (not in active map)
		for i := 0; i < 3; i++ {
			url := fmt.Sprintf("https://example.com/video%d", i)
			_, exists := handlers.active.Load(url)
			suite.False(exists, "URL %s should not be in active map", url)
		}
	})
}

// Test downloadHandler method

func (suite *TestEventHandlersSuite) TestDownloadHandler() {
	// helper to swap queue with buffered size 1 per subtest
	replaceQueue := func() {
		ch := make(chan domain.DownloadRequest, 1)
		suite.handlers.queue = ch
	}

	suite.Run("FailsRequestNotPassingValidation", func() {
		// Arrange
		replaceQueue()
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)
		validationErr := errors.New("validation failed")

		// Underlying Validate returns base error which handler wraps
		suite.downloader.EXPECT().Validate(suite.ctx, req).Return(validationErr)

		// Expect publication with StatusFailed and error containing substring
		suite.bus.EXPECT().Publish(mock.MatchedBy(func(ev domain.Event) bool {
			if ev.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp, ok := ev.Payload().(domain.DownloadResponse)
			if !ok {
				return false
			}
			if resp.Status != domain.StatusFailed {
				return false
			}
			if resp.Error == nil {
				return false
			}
			return strings.Contains(resp.Error.Error(), "validation failed")
		})).Return()

		// Act
		handler := suite.handlers.downloadHandler(suite.ctx)
		handler(event)
	})

	suite.Run("FailsWithErrDownloadBusyWhenNoFreeWorkers", func() {
		// Arrange
		replaceQueue()
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)

		// Fill queue (capacity 1)
		suite.handlers.queue <- domain.DownloadRequest{ID: "dummy", Url: "https://dummy"}
		// Mock successful validation so busy branch triggers
		suite.downloader.EXPECT().Validate(suite.ctx, req).Return(nil)
		suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusFailed, nil, domain.ErrDownloadBusy)).Return()

		// Act
		handler := suite.handlers.downloadHandler(suite.ctx)
		handler(event)
	})

	suite.Run("EnqueuesRequestWhenWorkersAvailable", func() {
		// Arrange
		replaceQueue()
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)
		suite.downloader.EXPECT().Validate(suite.ctx, req).Return(nil)
		suite.bus.EXPECT().Publish(mock.MatchedBy(func(ev domain.Event) bool {
			if ev.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp, ok := ev.Payload().(domain.DownloadResponse)
			if !ok {
				return false
			}
			return resp.Status == domain.StatusPending && resp.Request.ID == req.ID
		})).Return()

		// Act
		handler := suite.handlers.downloadHandler(suite.ctx)
		handler(event)

		// Assert request present in queue
		select {
		case receivedReq := <-suite.handlers.queue:
			suite.Equal(req.ID, receivedReq.ID)
		case <-time.After(100 * time.Millisecond):
			suite.Fail("Request not enqueued")
		}
	})

	suite.Run("CreatesEventWhenRequestEnqueued", func() {
		// Arrange
		replaceQueue()
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)
		suite.downloader.EXPECT().Validate(suite.ctx, req).Return(nil)
		suite.bus.EXPECT().Publish(mock.MatchedBy(func(ev domain.Event) bool {
			if ev.Type() != domain.DownloadResponseEvent {
				return false
			}
			resp, ok := ev.Payload().(domain.DownloadResponse)
			if !ok {
				return false
			}
			return resp.Status == domain.StatusPending && resp.Request.ID == req.ID && resp.Error == nil && resp.Episode == nil
		})).Return()

		// Act
		handler := suite.handlers.downloadHandler(suite.ctx)
		handler(event)
	})
}

// Test downloadValidate method

func (suite *TestEventHandlersSuite) TestDownloadValidate() {
	req := suite.createDownloadRequest()

	suite.Run("Success", func() {
		// Arrange
		suite.downloader.EXPECT().Validate(suite.ctx, req).Return(nil)

		// Act
		err := suite.handlers.downloadValidate(suite.ctx, req)

		// Assert
		suite.NoError(err)
	})

	suite.Run("DownloaderValidationFails", func() {
		// Arrange
		expectedErr := errors.New("validation failed")
		suite.downloader.EXPECT().Validate(suite.ctx, req).Return(expectedErr)

		// Act
		err := suite.handlers.downloadValidate(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "validation failed")
	})

	suite.Run("AlreadyDownloading", func() {
		// Arrange - mark URL as being downloaded
		suite.handlers.active.Store(req.Url, struct{}{})
		defer suite.handlers.active.Delete(req.Url) // cleanup

		// Act
		err := suite.handlers.downloadValidate(suite.ctx, req)

		// Assert
		suite.Equal(domain.ErrDownloadInProgress, err)
	})
}

// Test download method

func (suite *TestEventHandlersSuite) TestDownload() {
	req := suite.createDownloadRequest()

	suite.Run("Success", func() {
		// Arrange
		episode := suite.createEpisode()
		suite.downloader.EXPECT().Download(suite.ctx, req).Return(episode, nil)
		suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusSuccess, episode, nil)).Return()

		// Act
		suite.handlers.download(suite.ctx, req)

		// Assert
		// Verify URL is not in active map after completion
		_, exists := suite.handlers.active.Load(req.Url)
		suite.False(exists)
	})

	suite.Run("DownloadFails", func() {
		// Arrange
		downloadErr := errors.New("download failed")
		suite.downloader.EXPECT().Download(suite.ctx, req).Return(nil, downloadErr)
		suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusFailed, nil, downloadErr)).Return()

		// Act
		suite.handlers.download(suite.ctx, req)

		// Assert
		// Verify URL is not in active map after failure
		_, exists := suite.handlers.active.Load(req.Url)
		suite.False(exists)
	})

	suite.Run("AlreadyDownloading", func() {
		// Arrange - mark URL as being downloaded
		suite.handlers.active.Store(req.Url, struct{}{})
		suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusFailed, nil, domain.ErrDownloadInProgress)).Return()

		// Act
		suite.handlers.download(suite.ctx, req)

		// Assert - URL should still be in active map since this was a duplicate
		_, exists := suite.handlers.active.Load(req.Url)
		suite.True(exists)

		// Clean up for other tests
		suite.handlers.active.Delete(req.Url)
	})
}

// Test failRequest method

func (suite *TestEventHandlersSuite) TestFailRequest() {
	suite.Run("DownloadRequest", func() {
		// Arrange
		req := suite.createDownloadRequest()
		err := errors.New("test error")
		suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusFailed, nil, err)).Return()

		// Act
		suite.handlers.failRequest(req, err)

		// Assert - verify mock was called
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("BuildRequest", func() {
		// Arrange
		req := suite.createBuildRequest()
		err := errors.New("test error")
		suite.bus.EXPECT().Publish(suite.matchBuildResponseEvent(domain.StatusFailed, err)).Return()

		// Act
		suite.handlers.failRequest(req, err)

		// Assert - verify mock was called
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("UnknownRequestType", func() {
		// Arrange - use an unknown request type
		unknownReq := "unknown request"
		err := errors.New("test error")

		// Act - should not panic and should not publish any events
		suite.handlers.failRequest(unknownReq, err)

		// Assert - no event should be published (no EXPECT calls)
	})
}

// Test buildHandler

func (suite *TestEventHandlersSuite) TestBuildHandler() {
	suite.Run("Success", func() {
		// Arrange
		req := suite.createBuildRequest()
		event := domain.NewBuildRequestEvent(req)
		suite.builder.EXPECT().Build(suite.ctx).Return(nil)
		suite.bus.EXPECT().Publish(suite.matchBuildResponseEvent(domain.StatusSuccess, nil)).Return()

		// Act
		handler := suite.handlers.buildHandler(suite.ctx)
		handler(event)

		// Assert
		suite.builder.AssertExpectations(suite.T())
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("BuildFails", func() {
		// Arrange
		req := suite.createBuildRequest()
		event := domain.NewBuildRequestEvent(req)
		buildErr := errors.New("build failed")
		suite.builder.EXPECT().Build(suite.ctx).Return(buildErr)
		suite.bus.EXPECT().Publish(suite.matchBuildResponseEvent(domain.StatusFailed, buildErr)).Return()

		// Act
		handler := suite.handlers.buildHandler(suite.ctx)
		handler(event)

		// Assert
		suite.builder.AssertExpectations(suite.T())
		suite.bus.AssertExpectations(suite.T())
	})
}

// Test downloadWorker

func (suite *TestEventHandlersSuite) TestDownloadWorker() {
	suite.Run("ProcessesRequests", func() {
		// Arrange
		req := suite.createDownloadRequest()
		episode := suite.createEpisode()

		suite.downloader.EXPECT().Download(suite.ctx, req).Return(episode, nil)
		suite.bus.EXPECT().Publish(suite.matchDownloadResponseEvent(domain.StatusSuccess, episode, nil)).Return()

		// Start worker in background
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			suite.handlers.downloadWorker(suite.ctx, 1)
		}()

		// Act - send request to queue
		suite.handlers.queue <- req

		// Give some time for processing
		time.Sleep(50 * time.Millisecond)

		// Cancel context to stop worker
		suite.cancel()

		// Wait for worker to finish
		wg.Wait()

		// Assert - verify URL is not in active map
		_, exists := suite.handlers.active.Load(req.Url)
		suite.False(exists)
	})

	suite.Run("StopsOnContextCancel", func() {
		// Start worker in background
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			suite.handlers.downloadWorker(suite.ctx, 1)
		}()

		// Act - cancel context immediately
		suite.cancel()

		// Assert - worker should stop
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Expected - worker stopped
		case <-time.After(1 * time.Second):
			suite.Fail("Worker did not stop within timeout")
		}
	})
}

// Test logging handlers

func (suite *TestEventHandlersSuite) TestLoggingHandlers() {
	suite.Run("LogDownloadRequestHandler", func() {
		// Arrange
		req := suite.createDownloadRequest()
		event := domain.NewDownloadRequestEvent(req)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.logDownloadRequestHandler(event)
		})
	})

	suite.Run("LogDownloadResponseHandler", func() {
		// Arrange
		req := suite.createDownloadRequest()
		resp := domain.DownloadResponse{
			Status:  domain.StatusSuccess,
			Episode: suite.createEpisode(),
			Request: req,
		}
		event := domain.NewDownloadResponseEvent(resp)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.logDownloadResponseHandler(event)
		})
	})

	suite.Run("LogBuildRequestHandler", func() {
		// Arrange
		req := suite.createBuildRequest()
		event := domain.NewBuildRequestEvent(req)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.logBuildRequestHandler(event)
		})
	})

	suite.Run("LogBuildResponseHandler", func() {
		// Arrange
		req := suite.createBuildRequest()
		resp := domain.BuildResponse{
			Status:  domain.StatusSuccess,
			Request: req,
		}
		event := domain.NewBuildResponseEvent(resp)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.logBuildResponseHandler(event)
		})
	})

	suite.Run("LogFeedInfoRequestHandler", func() {
		// Arrange
		// create a feed info request and event
		req := domain.FeedInfoRequest{
			ID: "fi-1",
			Source: domain.RequestSource{
				UserID:    123,
				ChatID:    456,
				MessageID: 1,
			},
		}
		event := domain.NewFeedInfoRequestEvent(req)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.logFeedInfoRequestHandler(event)
		})
	})

	suite.Run("LogFeedInfoResponseHandler", func() {
		// Arrange
		req := domain.FeedInfoRequest{ID: "fi-1", Source: domain.RequestSource{UserID: 123}}
		resp := domain.FeedInfoResponse{
			Status:   domain.StatusSuccess,
			FeedInfo: &domain.FeedInfo{Title: "Test Feed"},
			Request:  req,
		}
		event := domain.NewFeedInfoResponseEvent(resp)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.logFeedInfoResponseHandler(event)
		})
	})
}

// Helper matchers for mock expectations

func (suite *TestEventHandlersSuite) matchDownloadResponseEvent(status domain.ResponseStatus, episode *domain.Episode, err error) interface{} {
	return mock.MatchedBy(func(event domain.Event) bool {
		if event.Type() != domain.DownloadResponseEvent {
			return false
		}
		resp, ok := event.Payload().(domain.DownloadResponse)
		if !ok {
			return false
		}

		if resp.Status != status {
			return false
		}

		if err != nil && (resp.Error == nil || resp.Error.Error() != err.Error()) {
			return false
		}

		if err == nil && resp.Error != nil {
			return false
		}

		if episode != nil && (resp.Episode == nil || resp.Episode.ID != episode.ID) {
			return false
		}

		if episode == nil && resp.Episode != nil {
			return false
		}

		return true
	})
}

func (suite *TestEventHandlersSuite) matchBuildResponseEvent(status domain.ResponseStatus, err error) interface{} {
	return mock.MatchedBy(func(event domain.Event) bool {
		if event.Type() != domain.BuildResponseEvent {
			return false
		}
		resp, ok := event.Payload().(domain.BuildResponse)
		if !ok {
			return false
		}

		if resp.Status != status {
			return false
		}

		if err != nil && (resp.Error == nil || resp.Error.Error() != err.Error()) {
			return false
		}

		if err == nil && resp.Error != nil {
			return false
		}

		return true
	})
}
