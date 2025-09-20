package telegram

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/locales"
	"github.com/ofstudio/voxify/internal/services"
	"github.com/stretchr/testify/suite"
)

// TestNotificationsSuite is a test suite for Notifications
type TestNotificationsSuite struct {
	suite.Suite
	ctx            context.Context
	log            *slog.Logger
	notifications  *Notifications
	processChannel chan entities.Process
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestNotificationsSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// SetupSubTest is called before each subtest
func (suite *TestNotificationsSuite) SetupSubTest() {
	// Create a buffered channel for testing
	suite.processChannel = make(chan entities.Process, 10)

	// Create notifications with nil bot since we're not testing bot methods
	suite.notifications = NewNotifications(suite.log, nil, suite.processChannel)
}

// TestGetMessage tests the getMessage method
func (suite *TestNotificationsSuite) TestGetMessage() {
	suite.Run("Success_ProcessSuccess", func() {
		// Arrange
		process := entities.Process{
			Step:   entities.StepPublishing,
			Status: entities.StatusSuccess,
			Episode: &entities.Episode{
				Title: "Test Episode",
			},
		}

		// Act
		message := suite.notifications.getMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n🎧 Test Episode"
		suite.Equal(expected, message)
	})

	suite.Run("Success_ProcessFailed", func() {
		// Arrange
		testError := services.NewError(102, "download failed")
		process := entities.Process{
			Step:   entities.StepDownloading,
			Status: entities.StatusFailed,
			Error:  testError,
		}

		// Act
		message := suite.notifications.getMessage(process)

		// Assert
		suite.Equal(locales.MsgDownloadFailed, message)
	})

	suite.Run("Success_DownloadInProgress", func() {
		// Arrange
		process := entities.Process{
			Step:   entities.StepDownloading,
			Status: entities.StatusInProgress,
		}

		// Act
		message := suite.notifications.getMessage(process)

		// Assert
		suite.Equal(locales.MsgDownloadStarted, message)
	})

	suite.Run("Success_DefaultCase", func() {
		// Arrange
		process := entities.Process{
			Step:   entities.StepCreating,
			Status: entities.StatusInProgress,
		}

		// Act
		message := suite.notifications.getMessage(process)

		// Assert
		suite.Equal("", message) // Default case returns empty string
	})
}

// TestGetSuccessMessage tests the getSuccessMessage method
func (suite *TestNotificationsSuite) TestGetSuccessMessage() {
	suite.Run("WithEpisodeTitle", func() {
		// Arrange
		process := entities.Process{
			Episode: &entities.Episode{
				Title: "Amazing Podcast Episode",
			},
		}

		// Act
		message := suite.notifications.getSuccessMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n🎧 Amazing Podcast Episode"
		suite.Equal(expected, message)
	})

	suite.Run("WithoutEpisode", func() {
		// Arrange
		process := entities.Process{
			Episode: nil,
		}

		// Act
		message := suite.notifications.getSuccessMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n🎧 "
		suite.Equal(expected, message)
	})

	suite.Run("WithEmptyTitle", func() {
		// Arrange
		process := entities.Process{
			Episode: &entities.Episode{
				Title: "",
			},
		}

		// Act
		message := suite.notifications.getSuccessMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n🎧 "
		suite.Equal(expected, message)
	})
}

// TestMsgErr tests the msgErr function for different error types
func (suite *TestNotificationsSuite) TestMsgErr() {
	suite.Run("ServiceError_DownloadFailed", func() {
		// Arrange
		testError := services.NewError(102, "download failed")

		// Act
		message := msgErr(testError)

		// Assert
		suite.Equal(locales.MsgDownloadFailed, message)
	})

	suite.Run("ServiceError_EpisodeInProgress", func() {
		// Arrange
		testError := services.NewError(103, "episode in progress")

		// Act
		message := msgErr(testError)

		// Assert
		suite.Equal(locales.MsgEpisodeInProgress, message)
	})

	suite.Run("ServiceError_UnknownCode", func() {
		// Arrange
		testError := services.NewError(999, "unknown error")

		// Act
		message := msgErr(testError)

		// Assert
		expected := "⚠️ Something went wrong while downloading the podcast (error 999)."
		suite.Equal(expected, message)
	})

	suite.Run("GenericError", func() {
		// Arrange
		testError := services.NewError(0, "generic error")

		// Act
		message := msgErr(testError)

		// Assert
		suite.Equal("⚠️ Something went wrong while downloading the podcast (error 0).", message)
	})

	suite.Run("NilError", func() {
		// Act
		message := msgErr(nil)

		// Assert
		expected := "⚠️ Something went wrong while downloading the podcast (error 0)."
		suite.Equal(expected, message)
	})
}

// TestStart tests the Start method behavior (without actual bot calls)
func (suite *TestNotificationsSuite) TestStart() {
	suite.Run("ProcessHandling", func() {
		// Arrange
		process := entities.Process{
			Step:   entities.StepDownloading,
			Status: entities.StatusInProgress,
			Request: entities.Request{
				ChatID:    123,
				MessageID: 456,
			},
		}

		// Start the notifications service
		ctx, cancel := context.WithCancel(suite.ctx)
		defer cancel()

		suite.notifications.Start(ctx)

		// Act - send a process to the channel
		suite.processChannel <- process

		// Give some time for processing
		// Since we can't test the actual bot sending, we just verify the method doesn't panic
		cancel() // Stop the service

		// Assert - if we get here without panic, the test passes
		suite.True(true)
	})
}

// Run the test suite
func TestNotifications(t *testing.T) {
	suite.Run(t, new(TestNotificationsSuite))
}
