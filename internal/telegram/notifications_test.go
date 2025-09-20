package telegram

import (
	"context"
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
		message := suite.notifications.getMessage(process)

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
		message := suite.notifications.getMessage(process)

		// Assert
		suite.Equal(locales.MsgDownloadFailed, message)
	})

	suite.Run("Success_DownloadInProgress", func() {
		// Arrange
		process := &entities.Process{
			Step:   entities.StepDownloading,
			Status: entities.StatusInProgress,
		}

		// Act
		message := suite.notifications.getMessage(process)

		// Assert
		suite.Equal(locales.MsgDownloadStarted, message)
	})

	suite.Run("DefaultCase_UnsupportedStatusStepCombination", func() {
		// Arrange
		process := &entities.Process{
			Step:   entities.StepCreating,
			Status: entities.StatusInProgress,
		}

		// Act
		message := suite.notifications.getMessage(process)

		// Assert
		suite.Equal(locales.MsgSomethingWentWrong, message)
	})
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

// Run the test suite
func TestNotifications(t *testing.T) {
	suite.Run(t, new(TestNotificationsSuite))
}
