package telegram

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/locales"
	"github.com/ofstudio/voxify/internal/mocks"
	"github.com/ofstudio/voxify/internal/services"
)

// TestNotificationsSuite is a test suite for Notifications
type TestNotificationsSuite struct {
	suite.Suite
	ctx           context.Context
	log           *slog.Logger
	mockNotifier  *mocks.MockNotifier
	notifications *Notifications
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestNotificationsSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// SetupSubTest is called before each subtest
func (suite *TestNotificationsSuite) SetupSubTest() {
	suite.mockNotifier = mocks.NewMockNotifier(suite.T())

	// Create notifications with nil bot since we're not testing bot methods
	suite.notifications = NewNotifications(nil, suite.mockNotifier, suite.log)
}

// TestNewNotifications tests the constructor
func (suite *TestNotificationsSuite) TestNewNotifications() {
	// Arrange
	mockNotifier := mocks.NewMockNotifier(suite.T())
	log := slog.Default()

	// Act
	notifications := NewNotifications(nil, mockNotifier, log)

	// Assert
	suite.NotNil(notifications)
	suite.Equal(mockNotifier, notifications.notifier)
	suite.NotNil(notifications.log)
}

// TestGetMessage tests the getMessage method
func (suite *TestNotificationsSuite) TestGetMessage() {
	suite.Run("Success_DownloadStarted", func() {
		// Arrange
		process := &entities.Process{
			Step:   entities.StepDownloading,
			Status: entities.StatusInProgress,
		}

		// Act
		message := suite.notifications.getMessage(suite.ctx, process)

		// Assert
		suite.Equal(locales.MsgDownloadStarted, message)
	})

	suite.Run("Success_ProcessSuccess", func() {
		// Arrange
		process := &entities.Process{
			Step:   entities.StepPublishing,
			Status: entities.StatusSuccess,
			Episode: &entities.Episode{
				Title: "Test Episode",
			},
		}

		// Act
		message := suite.notifications.getMessage(suite.ctx, process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n📖 **Test Episode**"
		suite.Equal(expected, message)
	})

	suite.Run("Success_ProcessFailed", func() {
		// Arrange
		testError := services.NewError(102, "download failed")
		process := &entities.Process{
			Step:   entities.StepDownloading,
			Status: entities.StatusFailed,
			Error:  testError,
		}

		// Act
		message := suite.notifications.getMessage(suite.ctx, process)

		// Assert
		suite.Equal(locales.MsgDownloadFailed, message)
	})

	suite.Run("EmptyMessage_UnsupportedStatusStepCombination", func() {
		// Arrange
		process := &entities.Process{
			Step:   entities.StepCreating, // Not supported for notifications
			Status: entities.StatusInProgress,
		}

		// Act
		message := suite.notifications.getMessage(suite.ctx, process)

		// Assert
		suite.Equal("", message, "Should return empty string for unsupported combinations")
	})

}

// TestGetDownloadStartedMessage tests the getDownloadStartedMessage method
func (suite *TestNotificationsSuite) TestGetDownloadStartedMessage() {
	// Act
	message := suite.notifications.getDownloadStartedMessage()

	// Assert
	suite.Equal(locales.MsgDownloadStarted, message)
}

// TestGetSuccessMessage tests the getSuccessMessage method
func (suite *TestNotificationsSuite) TestGetSuccessMessage() {
	suite.Run("WithEpisodeTitle", func() {
		// Arrange
		process := &entities.Process{
			Episode: &entities.Episode{
				Title: "Amazing Podcast Episode",
			},
		}

		// Act
		message := suite.notifications.getSuccessMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n📖 **Amazing Podcast Episode**"
		suite.Equal(expected, message)
	})

	suite.Run("WithoutEpisode", func() {
		// Arrange
		process := &entities.Process{
			Episode: nil,
		}

		// Act
		message := suite.notifications.getSuccessMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n📖 ****"
		suite.Equal(expected, message)
	})

	suite.Run("WithEmptyTitle", func() {
		// Arrange
		process := &entities.Process{
			Episode: &entities.Episode{
				Title: "",
			},
		}

		// Act
		message := suite.notifications.getSuccessMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n📖 ****"
		suite.Equal(expected, message)
	})

	suite.Run("WithSpecialCharactersInTitle", func() {
		// Arrange
		process := &entities.Process{
			Episode: &entities.Episode{
				Title: "Episode with **bold** and *italic* & special chars",
			},
		}

		// Act
		message := suite.notifications.getSuccessMessage(process)

		// Assert
		expected := "✅ Podcast downloaded successfully!\n\n📖 **Episode with **bold** and *italic* & special chars**"
		suite.Equal(expected, message)
	})
}

// TestGetErrorMessage tests the getErrorMessage method
func (suite *TestNotificationsSuite) TestGetErrorMessage() {
	suite.Run("NoError", func() {
		// Arrange
		process := &entities.Process{
			Error: nil,
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgSomethingWentWrong, message)
	})

	suite.Run("ServicesError_NoMatchingPlatform", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(101, "no matching platform"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgNoMatchingPlatform, message)
	})

	suite.Run("ServicesError_DownloadFailed", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(102, "download failed"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgDownloadFailed, message)
	})

	suite.Run("ServicesError_EpisodeInProgress", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(103, "episode in progress"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgEpisodeInProgress, message)
	})

	suite.Run("ServicesError_EpisodeExists", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(104, "episode exists"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgEpisodeExists, message)
	})

	suite.Run("ServicesError_ProcessInterrupted", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(105, "process interrupted"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgProcessInterrupted, message)
	})

	suite.Run("ServicesError_ProcessUpsert", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(201, "failed to upsert process"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		expected := "⚠️ Something went wrong while downloading the podcast (error 201)."
		suite.Equal(expected, message)
	})

	suite.Run("ServicesError_ProcessGetByURL", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(202, "failed to get process by URL"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		expected := "⚠️ Something went wrong while downloading the podcast (error 202)."
		suite.Equal(expected, message)
	})

	suite.Run("ServicesError_DownloadDir", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(301, "download directory error"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		expected := "⚠️ Something went wrong while downloading the podcast (error 301)."
		suite.Equal(expected, message)
	})

	suite.Run("ServicesError_UnknownCode", func() {
		// Arrange
		process := &entities.Process{
			Error: services.NewError(999, "unknown error"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		expected := "⚠️ Something went wrong while downloading the podcast (error 999)."
		suite.Equal(expected, message)
	})

	suite.Run("NonServicesError", func() {
		// Arrange
		process := &entities.Process{
			Error: errors.New("generic error"),
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgSomethingWentWrong, message)
	})

	suite.Run("NilProcessWithError", func() {
		// Arrange
		process := &entities.Process{
			Error: nil,
		}

		// Act
		message := suite.notifications.getErrorMessage(process)

		// Assert
		suite.Equal(locales.MsgSomethingWentWrong, message)
	})
}

// TestGetErrorMessageByCode tests the getErrorMessageByCode method
func (suite *TestNotificationsSuite) TestGetErrorMessageByCode() {
	testCases := []struct {
		name     string
		code     int
		expected string
	}{
		{"NoMatchingPlatform", 101, locales.MsgNoMatchingPlatform},
		{"DownloadFailed", 102, locales.MsgDownloadFailed},
		{"EpisodeInProgress", 103, locales.MsgEpisodeInProgress},
		{"EpisodeExists", 104, locales.MsgEpisodeExists},
		{"ProcessInterrupted", 105, locales.MsgProcessInterrupted},
		{"ProcessUpsertError", 201, "⚠️ Something went wrong while downloading the podcast (error 201)."},
		{"ProcessGetByURLError", 202, "⚠️ Something went wrong while downloading the podcast (error 202)."},
		{"EpisodeGetByOriginalURLError", 203, "⚠️ Something went wrong while downloading the podcast (error 203)."},
		{"DownloadDirError", 301, "⚠️ Something went wrong while downloading the podcast (error 301)."},
		{"MoveFileError", 303, "⚠️ Something went wrong while downloading the podcast (error 303)."},
		{"UnknownCode_999", 999, "⚠️ Something went wrong while downloading the podcast (error 999)."},
		{"UnknownCode_0", 0, "⚠️ Something went wrong while downloading the podcast (error 0)."},
		{"NegativeCode", -1, "⚠️ Something went wrong while downloading the podcast (error -1)."},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Act
			message := suite.notifications.getErrorMessageByCode(tc.code)

			// Assert
			suite.Equal(tc.expected, message)
		})
	}
}

// Run the test suite
func TestNotifications(t *testing.T) {
	suite.Run(t, new(TestNotificationsSuite))
}
