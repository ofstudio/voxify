package handlers

import (
	"context"
	"log/slog"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/mocks"
	"github.com/ofstudio/voxify/internal/templates"
)

// TestTelegramHandlers is the entry point for running the test suite
func TestTelegramHandlers(t *testing.T) {
	suite.Run(t, new(TestTelegramHandlersSuite))
}

// TestTelegramHandlersSuite is a test suite for TelegramHandlers
type TestTelegramHandlersSuite struct {
	suite.Suite
	ctx      context.Context
	cancel   context.CancelFunc
	log      *slog.Logger
	bus      *mocks.MockEventBus
	bot      *mocks.MockBot
	api      *mocks.MockAPI
	handlers *TelegramHandlers
}

// SetupTest is called once before each test start
func (suite *TestTelegramHandlersSuite) SetupTest() {
	suite.SetupSubTest()
}

// SetupSubTest is called before each subtest in the suite
func (suite *TestTelegramHandlersSuite) SetupSubTest() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.log = slog.Default()
	suite.bus = mocks.NewMockEventBus(suite.T())
	suite.bot = mocks.NewMockBot(suite.T())
	suite.api = mocks.NewMockAPI(suite.T())

	suite.handlers = NewTelegramHandlers(suite.log, suite.bus, suite.bot)
}

// TearDownTest is called after each test in the suite completes
func (suite *TestTelegramHandlersSuite) TearDownTest() {
	suite.TearDownSubTest()
}

// TearDownSubTest is called after each subtest in the suite
func (suite *TestTelegramHandlersSuite) TearDownSubTest() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// Helper functions

func (suite *TestTelegramHandlersSuite) createMessage(text string, userID int64, chatID int64, messageID int) *models.Message {
	return &models.Message{
		ID:   messageID,
		Text: text,
		From: &models.User{
			ID: userID,
		},
		Chat: models.Chat{
			ID: chatID,
		},
	}
}

func (suite *TestTelegramHandlersSuite) createUpdate(message *models.Message) *models.Update {
	return &models.Update{
		Message: message,
	}
}

func (suite *TestTelegramHandlersSuite) createStartUpdate() *models.Update {
	message := suite.createMessage("/start", 12345, 67890, 111)
	return suite.createUpdate(message)
}

func (suite *TestTelegramHandlersSuite) createBuildUpdate() *models.Update {
	message := suite.createMessage("/build", 12345, 67890, 222)
	return suite.createUpdate(message)
}

func (suite *TestTelegramHandlersSuite) createUrlUpdate() *models.Update {
	message := suite.createMessage("https://example.com/video", 12345, 67890, 333)
	return suite.createUpdate(message)
}

// Test NewTelegramHandlers

func (suite *TestTelegramHandlersSuite) TestNewTelegramHandlers() {
	// Act
	handlers := NewTelegramHandlers(suite.log, suite.bus, suite.bot)

	// Assert
	suite.NotNil(handlers)
	suite.Equal(suite.log, handlers.log)
	suite.Equal(suite.bus, handlers.bus)
	suite.Equal(suite.bot, handlers.bot)
}

// Test ErrorsHandler method

func (suite *TestTelegramHandlersSuite) TestErrorsHandler() {
	suite.Run("HandlesError", func() {
		// Arrange
		testError := domain.ErrDownloadFailed

		// Act
		errorHandler := suite.handlers.ErrorsHandler()

		// Assert - should not panic
		suite.NotPanics(func() {
			errorHandler(testError)
		})
	})
}

// Test CmdStartHandler method

func (suite *TestTelegramHandlersSuite) TestHandleStart() {
	suite.Run("SendsWelcomeMessage", func() {
		// Arrange
		update := suite.createStartUpdate()
		expectedParams := &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      templates.MsgStart,
			ParseMode: models.ParseModeHTML,
		}

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(expectedParams)).
			Return(&models.Message{}, nil)

		// Act
		handler := suite.handlers.CmdStartHandler()
		handler(suite.ctx, suite.api, update)

		// Assert
		suite.api.AssertExpectations(suite.T())
	})

	suite.Run("IgnoresUpdateWithoutMessage", func() {
		// Arrange
		update := &models.Update{Message: nil}

		// Act - should not panic and should not call api methods
		handler := suite.handlers.CmdStartHandler()
		handler(suite.ctx, suite.api, update)

		// Assert - no api methods should be called
	})

	suite.Run("HandlesSendMessageError", func() {
		// Arrange
		update := suite.createStartUpdate()
		expectedParams := &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      templates.MsgStart,
			ParseMode: models.ParseModeHTML,
		}

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(expectedParams)).
			Return(nil, domain.ErrDownloadFailed)

		// Act - should not panic even if SendMessage fails
		handler := suite.handlers.CmdStartHandler()
		suite.NotPanics(func() {
			handler(suite.ctx, suite.api, update)
		})

		// Assert
		suite.api.AssertExpectations(suite.T())
	})
}

// Test CmdBuildHandler method

func (suite *TestTelegramHandlersSuite) TestHandleBuild() {
	suite.Run("PublishesBuildRequestEvent", func() {
		// Arrange
		update := suite.createBuildUpdate()

		expectedSource := domain.RequestSource{
			UserID:    update.Message.From.ID,
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
		}

		suite.bus.EXPECT().Publish(suite.matchBuildRequestEvent(expectedSource)).Return()

		// Act
		handler := suite.handlers.CmdBuildHandler()
		handler(suite.ctx, suite.api, update)

		// Assert
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("IgnoresUpdateWithoutMessage", func() {
		// Arrange
		update := &models.Update{Message: nil}

		// Act - should not panic and should not publish events
		handler := suite.handlers.CmdBuildHandler()
		handler(suite.ctx, suite.api, update)

		// Assert - no events should be published
	})
}

// Test UrlHandler method

func (suite *TestTelegramHandlersSuite) TestHandleUrl() {
	suite.Run("PublishesDownloadRequestEvent", func() {
		// Arrange
		update := suite.createUrlUpdate()

		expectedSource := domain.RequestSource{
			UserID:    update.Message.From.ID,
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
		}

		suite.bus.EXPECT().Publish(suite.matchDownloadRequestEvent(expectedSource, update.Message.Text)).Return()

		// Act
		handler := suite.handlers.UrlHandler()
		handler(suite.ctx, suite.api, update)

		// Assert
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("IgnoresUpdateWithoutMessage", func() {
		// Arrange
		update := &models.Update{Message: nil}

		// Act - should not panic and should not publish events
		handler := suite.handlers.UrlHandler()
		handler(suite.ctx, suite.api, update)

		// Assert - no events should be published
	})

	suite.Run("IgnoresUpdateWithEmptyText", func() {
		// Arrange
		message := suite.createMessage("", 12345, 67890, 333)
		update := suite.createUpdate(message)

		// Act - should not panic and should not publish events
		handler := suite.handlers.UrlHandler()
		handler(suite.ctx, suite.api, update)

		// Assert - no events should be published
	})
}

// Test requestSource method

func (suite *TestTelegramHandlersSuite) TestRequestSource() {
	suite.Run("CreatesCorrectRequestSource", func() {
		// Arrange
		message := suite.createMessage("test", 12345, 67890, 111)

		// Act
		source := suite.handlers.requestSource(message)

		// Assert
		expectedSource := domain.RequestSource{
			UserID:    12345,
			ChatID:    67890,
			MessageID: 111,
		}
		suite.Equal(expectedSource, source)
	})
}

// Test requestID method

func (suite *TestTelegramHandlersSuite) TestRequestID() {
	suite.Run("GeneratesNonEmptyID", func() {
		// Act
		id := suite.handlers.requestID()

		// Assert
		suite.NotEmpty(id)
		suite.Len(id, 10) // randtoken.New(10) should generate 10-character string
	})

	suite.Run("GeneratesUniqueIDs", func() {
		// Act
		id1 := suite.handlers.requestID()
		id2 := suite.handlers.requestID()

		// Assert
		suite.NotEqual(id1, id2)
	})
}

// Test sendMessage method

func (suite *TestTelegramHandlersSuite) TestSendMessage() {
	suite.Run("SendsMessageSuccessfully", func() {
		// Arrange
		params := &bot.SendMessageParams{
			ChatID: 67890,
			Text:   "test message",
		}
		expectedMessage := &models.Message{ID: 123}

		suite.api.EXPECT().SendMessage(suite.ctx, params).Return(expectedMessage, nil)

		// Act
		suite.handlers.sendMessage(suite.ctx, suite.api, params)

		// Assert
		suite.api.AssertExpectations(suite.T())
	})

	suite.Run("HandlesErrorGracefully", func() {
		// Arrange
		params := &bot.SendMessageParams{
			ChatID: 67890,
			Text:   "test message",
		}

		suite.api.EXPECT().SendMessage(suite.ctx, params).Return(nil, domain.ErrDownloadFailed)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.sendMessage(suite.ctx, suite.api, params)
		})

		// Assert
		suite.api.AssertExpectations(suite.T())
	})
}

// Helper methods for matching mock expectations

func (suite *TestTelegramHandlersSuite) matchSendMessageParams(expected *bot.SendMessageParams) interface{} {
	return mock.MatchedBy(func(params *bot.SendMessageParams) bool {
		return params.ChatID == expected.ChatID &&
			params.Text == expected.Text &&
			params.ParseMode == expected.ParseMode
	})
}

func (suite *TestTelegramHandlersSuite) matchBuildRequestEvent(expectedSource domain.RequestSource) interface{} {
	return mock.MatchedBy(func(event domain.Event) bool {
		if event.Type() == domain.BuildRequestEvent {
			if buildRequest, ok := event.Payload().(domain.BuildRequest); ok {
				return buildRequest.Source != nil &&
					buildRequest.Source.UserID == expectedSource.UserID &&
					buildRequest.Source.ChatID == expectedSource.ChatID &&
					buildRequest.Source.MessageID == expectedSource.MessageID
			}
		}
		return false
	})
}

func (suite *TestTelegramHandlersSuite) matchDownloadRequestEvent(expectedSource domain.RequestSource, expectedUrl string) interface{} {
	return mock.MatchedBy(func(event domain.Event) bool {
		if event.Type() == domain.DownloadRequestEvent {
			if downloadRequest, ok := event.Payload().(domain.DownloadRequest); ok {
				return downloadRequest.Source.UserID == expectedSource.UserID &&
					downloadRequest.Source.ChatID == expectedSource.ChatID &&
					downloadRequest.Source.MessageID == expectedSource.MessageID &&
					downloadRequest.Url == expectedUrl
			}
		}
		return false
	})
}
