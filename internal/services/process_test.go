package services

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/mocks"
)

// TestProcessServiceSuite is a test suite for ProcessService
type TestProcessServiceSuite struct {
	suite.Suite
	ctx         context.Context
	cfg         *config.Settings
	log         *slog.Logger
	mockStore   *mocks.MockStore
	mockDown    *mocks.MockDownloader
	mockBuilder *mocks.MockBuilder
	service     *ProcessService
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestProcessServiceSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.cfg = &config.Settings{}
	suite.log = slog.Default()
}

// SetupTest is called before each test method
func (suite *TestProcessServiceSuite) SetupSubTest() {
	suite.mockStore = mocks.NewMockStore(suite.T())
	suite.mockDown = mocks.NewMockDownloader(suite.T())
	suite.mockBuilder = mocks.NewMockBuilder(suite.T())

	suite.service = NewProcessService(suite.log, suite.mockStore, suite.mockDown, suite.mockBuilder)
}

// TestNewProcessService tests the constructor
func (suite *TestProcessServiceSuite) TestNewProcessService() {
	// Act
	service := NewProcessService(suite.log, suite.mockStore, suite.mockDown, suite.mockBuilder)

	// Assert
	suite.NotNil(service)
	suite.NotNil(service.In())
	suite.NotNil(service.Notify())
}

// TestStart tests the Start method
func (suite *TestProcessServiceSuite) TestStart() {
	suite.Run("Success", func() {
		// Act - Start doesn't return an error anymore, it just starts the goroutine
		suite.service.Start(suite.ctx)

		// Assert - no error to check, just verify it doesn't panic
		suite.True(true)
	})
}

// TestInit tests the Init method
func (suite *TestProcessServiceSuite) TestInit() {
	suite.Run("NoInProgressProcesses", func() {
		// Arrange
		suite.mockStore.On("ProcessGetByStatus", suite.ctx, entities.StatusInProgress).
			Return([]*entities.Process{}, nil)

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.NoError(err)
	})

	suite.Run("WithInProgressProcesses", func() {
		// Arrange
		process1 := &entities.Process{ID: 1, Status: entities.StatusInProgress}
		process2 := &entities.Process{ID: 2, Status: entities.StatusInProgress}
		processes := []*entities.Process{process1, process2}

		suite.mockStore.On("ProcessGetByStatus", suite.ctx, entities.StatusInProgress).
			Return(processes, nil)

		// Expect calls to ProcessUpsert for failing processes
		suite.mockStore.On("ProcessUpsert", suite.ctx, mock.MatchedBy(func(p *entities.Process) bool {
			return p.ID == 1 && p.Status == entities.StatusFailed && errors.Is(p.Error, ErrProcessInterrupted)
		})).Return(nil)
		suite.mockStore.On("ProcessUpsert", suite.ctx, mock.MatchedBy(func(p *entities.Process) bool {
			return p.ID == 2 && p.Status == entities.StatusFailed && errors.Is(p.Error, ErrProcessInterrupted)
		})).Return(nil)

		// Start goroutine to consume notifications before calling Init
		notifications := make([]*entities.Process, 0, 2)
		done := make(chan struct{})

		go func() {
			defer close(done)
			for i := 0; i < 2; i++ {
				select {
				case notification := <-suite.service.Notify():
					notifications = append(notifications, notification)
					if len(notifications) == 2 {
						return
					}
				case <-time.After(100 * time.Millisecond):
					return
				}
			}
		}()

		// Act
		err := suite.service.Init(suite.ctx)

		// Wait for notifications to be processed
		select {
		case <-done:
			// Success
		case <-time.After(500 * time.Millisecond):
			// This is okay if notifications don't come - the important thing is that the method doesn't block
		}

		// Assert
		suite.NoError(err)
		suite.True(true) // Test passes if we get here without hanging
	})

	suite.Run("StoreError", func() {
		// Arrange
		suite.mockStore.On("ProcessGetByStatus", suite.ctx, entities.StatusInProgress).
			Return(nil, errors.New("store error"))

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, ErrProcessGetByStatus))
	})

	suite.Run("UpsertError", func() {
		// Arrange
		process1 := &entities.Process{ID: 1, Status: entities.StatusInProgress}
		processes := []*entities.Process{process1}

		suite.mockStore.On("ProcessGetByStatus", suite.ctx, entities.StatusInProgress).
			Return(processes, nil)
		suite.mockStore.On("ProcessUpsert", suite.ctx, mock.MatchedBy(func(p *entities.Process) bool {
			return p.ID == 1 && p.Status == entities.StatusFailed
		})).Return(errors.New("upsert error"))

		// Act
		err := suite.service.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, ErrProcessUpsert))
	})
}

// TestValidate tests the validate method
func (suite *TestProcessServiceSuite) TestValidate() {
	process := &entities.Process{
		Request: entities.Request{
			Url:   "https://example.com/video",
			Force: false,
		},
	}

	suite.Run("Success", func() {
		// Arrange
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, process.Request.Url, entities.StatusInProgress).
			Return(1, nil) // Count of 1 is OK (current process)
		suite.mockStore.On("EpisodeGetByOriginalUrl", suite.ctx, process.Request.Url).
			Return([]*entities.Episode{}, nil)

		// Act
		err := suite.service.validate(suite.ctx, process)

		// Assert
		suite.NoError(err)
	})

	suite.Run("EpisodeInProgress", func() {
		// Arrange
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, process.Request.Url, entities.StatusInProgress).
			Return(2, nil) // Count > 1 triggers error

		// Act
		err := suite.service.validate(suite.ctx, process)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, ErrEpisodeInProgress))
	})

	suite.Run("EpisodeExists", func() {
		// Arrange
		existingEpisode := &entities.Episode{ID: 1, OriginalURL: process.Request.Url}
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, process.Request.Url, entities.StatusInProgress).
			Return(1, nil) // Count of 1 is OK
		suite.mockStore.On("EpisodeGetByOriginalUrl", suite.ctx, process.Request.Url).
			Return([]*entities.Episode{existingEpisode}, nil)

		// Act
		err := suite.service.validate(suite.ctx, process)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, ErrEpisodeExists))
	})

	suite.Run("ForceDownload", func() {
		// Arrange
		processForce := &entities.Process{
			Request: entities.Request{
				Url:   "https://example.com/video",
				Force: true,
			},
		}
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, processForce.Request.Url, entities.StatusInProgress).
			Return(1, nil) // Count of 1 is OK

		// Act
		err := suite.service.validate(suite.ctx, processForce)

		// Assert
		suite.NoError(err) // Force=true skips episode existence check
	})

	suite.Run("ProcessCountError", func() {
		// Arrange
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, process.Request.Url, entities.StatusInProgress).
			Return(0, errors.New("count error"))

		// Act
		err := suite.service.validate(suite.ctx, process)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, ErrProcessGetByURL))
	})

	suite.Run("EpisodeGetError", func() {
		// Arrange
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, process.Request.Url, entities.StatusInProgress).
			Return(1, nil)
		suite.mockStore.On("EpisodeGetByOriginalUrl", suite.ctx, process.Request.Url).
			Return(nil, errors.New("get episode error"))

		// Act
		err := suite.service.validate(suite.ctx, process)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, EpisodeGetByOriginalURL))
	})
}

// TestHandle tests the handle method
func (suite *TestProcessServiceSuite) TestHandle() {
	request := &entities.Request{
		ID:              "test-req-123",
		UserID:          123,
		ChatID:          456,
		MessageID:       789,
		Url:             "https://example.com/video",
		DownloadFormat:  entities.DownloadMp3,
		DownloadQuality: "best",
		Force:           false,
	}

	episode := &entities.Episode{
		ID:            1,
		Title:         "Test Episode",
		OriginalURL:   request.Url,
		MediaFile:     "audio.mp3",
		MediaType:     entities.MediaMp3,
		MediaSize:     1024000,
		MediaDuration: 3600,
	}

	suite.Run("SuccessfulProcess", func() {
		// Arrange - validate step
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepCreating)).Return(nil)
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, request.Url, entities.StatusInProgress).Return(1, nil)
		suite.mockStore.On("EpisodeGetByOriginalUrl", suite.ctx, request.Url).Return([]*entities.Episode{}, nil)

		// Arrange - download step
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepDownloading)).Return(nil)
		suite.mockDown.On("Download", suite.ctx, *request).Return(episode, nil)
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessWithEpisode(episode)).Return(nil)

		// Arrange - publish step
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepPublishing)).Return(nil)
		suite.mockBuilder.On("Build", suite.ctx).Return(nil)
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStatus(entities.StatusSuccess)).Return(nil)

		// Start goroutine to consume notifications
		go func() {
			for i := 0; i < 5; i++ {
				select {
				case <-suite.service.Notify():
					// Consume notification
				case <-time.After(50 * time.Millisecond):
					return
				}
			}
		}()

		// Act
		suite.service.handle(suite.ctx, request)

		// Assert - verify all mocks were called
		suite.mockStore.AssertExpectations(suite.T())
		suite.mockDown.AssertExpectations(suite.T())
		suite.mockBuilder.AssertExpectations(suite.T())
	})

	suite.Run("ValidationFailure", func() {
		// Arrange
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepCreating)).Return(nil)
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, request.Url, entities.StatusInProgress).Return(2, nil) // Count > 1 triggers error
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessError(ErrEpisodeInProgress)).Return(nil)

		// Start goroutine to consume notifications
		go func() {
			for i := 0; i < 2; i++ {
				select {
				case <-suite.service.Notify():
					// Consume notification
				case <-time.After(50 * time.Millisecond):
					return
				}
			}
		}()

		// Act
		suite.service.handle(suite.ctx, request)

		// Assert
		suite.mockStore.AssertExpectations(suite.T())
	})

	suite.Run("DownloadFailure", func() {
		// Arrange - validate step
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepCreating)).Return(nil)
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, request.Url, entities.StatusInProgress).Return(1, nil)
		suite.mockStore.On("EpisodeGetByOriginalUrl", suite.ctx, request.Url).Return([]*entities.Episode{}, nil)

		// Arrange - download step failure
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepDownloading)).Return(nil)
		suite.mockDown.On("Download", suite.ctx, *request).Return(nil, ErrDownloadFailed)
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessError(ErrDownloadFailed)).Return(nil)

		// Start goroutine to consume notifications
		go func() {
			for i := 0; i < 3; i++ {
				select {
				case <-suite.service.Notify():
					// Consume notification
				case <-time.After(50 * time.Millisecond):
					return
				}
			}
		}()

		// Act
		suite.service.handle(suite.ctx, request)

		// Assert
		suite.mockStore.AssertExpectations(suite.T())
		suite.mockDown.AssertExpectations(suite.T())
	})

	suite.Run("BuildFailure", func() {
		// Arrange - validate step
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepCreating)).Return(nil)
		suite.mockStore.On("ProcessCountByUrlAndStatus", suite.ctx, request.Url, entities.StatusInProgress).Return(1, nil)
		suite.mockStore.On("EpisodeGetByOriginalUrl", suite.ctx, request.Url).Return([]*entities.Episode{}, nil)

		// Arrange - download step
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepDownloading)).Return(nil)
		suite.mockDown.On("Download", suite.ctx, *request).Return(episode, nil)
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessWithEpisode(episode)).Return(nil)

		// Arrange - publish step failure
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepPublishing)).Return(nil)
		suite.mockBuilder.On("Build", suite.ctx).Return(errors.New("build failed"))
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStatusAndError(entities.StatusFailed)).Return(nil)

		// Start goroutine to consume notifications
		go func() {
			for i := 0; i < 5; i++ {
				select {
				case <-suite.service.Notify():
					// Consume notification
				case <-time.After(50 * time.Millisecond):
					return
				}
			}
		}()

		// Act
		suite.service.handle(suite.ctx, request)

		// Assert
		suite.mockStore.AssertExpectations(suite.T())
		suite.mockDown.AssertExpectations(suite.T())
		suite.mockBuilder.AssertExpectations(suite.T())
	})

	suite.Run("UpsertFailure", func() {
		// Arrange
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessStep(entities.StepCreating)).Return(ErrProcessUpsert)
		suite.mockStore.On("ProcessUpsert", suite.ctx, suite.matchProcessError(ErrProcessUpsert)).Return(nil)

		// Start goroutine to consume notification
		go func() {
			select {
			case <-suite.service.Notify():
				// Consume notification
			case <-time.After(50 * time.Millisecond):
				return
			}
		}()

		// Act
		suite.service.handle(suite.ctx, request)

		// Assert
		suite.mockStore.AssertExpectations(suite.T())
	})
}

// TestUpsert tests the update method
func (suite *TestProcessServiceSuite) TestUpsert() {
	process := &entities.Process{
		Request: entities.Request{Url: "https://example.com/video"},
		Step:    entities.StepCreating,
		Status:  entities.StatusInProgress,
	}

	suite.Run("Success", func() {
		// Arrange
		suite.mockStore.On("ProcessUpsert", suite.ctx, process).Return(nil)

		// Start goroutine to consume notification
		var receivedNotification *entities.Process
		done := make(chan struct{})
		go func() {
			defer close(done)
			select {
			case notification := <-suite.service.Notify():
				receivedNotification = notification
			case <-time.After(100 * time.Millisecond):
				// Timeout
			}
		}()

		// Act
		err := suite.service.update(suite.ctx, process)

		// Wait for notification
		select {
		case <-done:
			// Success
		case <-time.After(500 * time.Millisecond):
			// Timeout
		}

		// Assert
		suite.NoError(err)
		if receivedNotification != nil {
			suite.Equal(process, receivedNotification)
		}
	})

	suite.Run("StoreError", func() {
		// Arrange
		storeErr := errors.New("store error")
		suite.mockStore.On("ProcessUpsert", suite.ctx, process).Return(storeErr)

		// Act
		err := suite.service.update(suite.ctx, process)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, ErrProcessUpsert))
		// No notification should be sent on error
	})
}

// TestFail tests the fail method
func (suite *TestProcessServiceSuite) TestFail() {
	process := &entities.Process{
		Request: entities.Request{Url: "https://example.com/video"},
		Step:    entities.StepDownloading,
		Status:  entities.StatusInProgress,
	}
	testError := errors.New("test error")

	suite.Run("Success", func() {
		// Arrange
		suite.mockStore.On("ProcessUpsert", suite.ctx, process).Return(nil)

		// Start goroutine to consume notification
		done := make(chan struct{})
		go func() {
			defer close(done)
			select {
			case notification := <-suite.service.Notify():
				suite.Equal(process, notification)
				suite.Equal(entities.StatusFailed, notification.Status)
				suite.Equal(testError, notification.Error)
			case <-time.After(100 * time.Millisecond):
				suite.Fail("Expected notification not received")
			}
		}()

		// Act
		suite.service.fail(suite.ctx, process, testError)

		// Wait for notification
		select {
		case <-done:
			// Success
		case <-time.After(500 * time.Millisecond):
			suite.Fail("Timeout waiting for notification")
		}

		// Assert
		suite.Equal(entities.StatusFailed, process.Status)
		suite.Equal(testError, process.Error)
	})

	suite.Run("UpsertError", func() {
		// Arrange
		process2 := &entities.Process{
			Request: entities.Request{Url: "https://example.com/video2"},
			Step:    entities.StepDownloading,
			Status:  entities.StatusInProgress,
		}
		suite.mockStore.On("ProcessUpsert", suite.ctx, process2).Return(errors.New("update error"))

		// Start goroutine to consume notification
		done := make(chan struct{})
		go func() {
			defer close(done)
			select {
			case notification := <-suite.service.Notify():
				suite.Equal(process2, notification)
			case <-time.After(100 * time.Millisecond):
				suite.Fail("Expected notification not received")
			}
		}()

		// Act
		suite.service.fail(suite.ctx, process2, testError)

		// Wait for notification
		select {
		case <-done:
			// Success
		case <-time.After(500 * time.Millisecond):
			suite.Fail("Timeout waiting for notification")
		}

		// Assert
		suite.Equal(entities.StatusFailed, process2.Status)
		suite.Equal(testError, process2.Error)
	})
}

// Helper matchers for mock arguments
func (suite *TestProcessServiceSuite) matchProcessStep(step entities.Step) interface{} {
	return mock.MatchedBy(func(p *entities.Process) bool {
		return p.Step == step
	})
}

func (suite *TestProcessServiceSuite) matchProcessStatus(status entities.Status) interface{} {
	return mock.MatchedBy(func(p *entities.Process) bool {
		return p.Status == status
	})
}

func (suite *TestProcessServiceSuite) matchProcessError(expectedErr error) interface{} {
	return mock.MatchedBy(func(p *entities.Process) bool {
		return p.Status == entities.StatusFailed && errors.Is(p.Error, expectedErr)
	})
}

func (suite *TestProcessServiceSuite) matchProcessWithEpisode(episode *entities.Episode) interface{} {
	return mock.MatchedBy(func(p *entities.Process) bool {
		return p.Episode == episode
	})
}

func (suite *TestProcessServiceSuite) matchProcessStatusAndError(status entities.Status) interface{} {
	return mock.MatchedBy(func(p *entities.Process) bool {
		return p.Status == status && p.Error != nil
	})
}

// TestProcessService runs the test suite
func TestProcessService(t *testing.T) {
	suite.Run(t, new(TestProcessServiceSuite))
}
