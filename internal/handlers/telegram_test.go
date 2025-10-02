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
	"github.com/ofstudio/voxify/pkg/telegram"
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

	suite.handlers = NewTelegramHandlers(suite.log, suite.bus)
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
	handlers := NewTelegramHandlers(suite.log, suite.bus)

	// Assert
	suite.NotNil(handlers)
	suite.Equal(suite.log, handlers.log)
	suite.Equal(suite.bus, handlers.bus)
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

		suite.bot.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(expectedParams)).
			Return(&models.Message{}, nil)

		// Act
		handler := suite.handlers.CmdStartHandler()
		handler(suite.ctx, suite.bot, update)

		// Assert
		suite.bot.AssertExpectations(suite.T())
	})

	suite.Run("IgnoresUpdateWithoutMessage", func() {
		// Arrange
		update := &models.Update{Message: nil}

		// Act - should not panic and should not call bot methods
		handler := suite.handlers.CmdStartHandler()
		handler(suite.ctx, suite.bot, update)

		// Assert - no bot methods should be called
	})

	suite.Run("HandlesSendMessageError", func() {
		// Arrange
		update := suite.createStartUpdate()
		expectedParams := &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      templates.MsgStart,
			ParseMode: models.ParseModeHTML,
		}

		suite.bot.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(expectedParams)).
			Return(nil, domain.ErrDownloadFailed)

		// Act - should not panic even if SendMessage fails
		handler := suite.handlers.CmdStartHandler()
		suite.NotPanics(func() {
			handler(suite.ctx, suite.bot, update)
		})

		// Assert
		suite.bot.AssertExpectations(suite.T())
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
		// CmdBuildHandler now returns telegram.HandlerFunc which expects telegram.Bot
		handler(suite.ctx, suite.bot, update)

		// Assert
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("IgnoresUpdateWithoutMessage", func() {
		// Arrange
		update := &models.Update{Message: nil}

		// Act - should not panic and should not publish events
		handler := suite.handlers.CmdBuildHandler()
		handler(suite.ctx, suite.bot, update)

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
		// UrlHandler now returns telegram.HandlerFunc which expects telegram.Bot
		handler(suite.ctx, suite.bot, update)

		// Assert
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("IgnoresUpdateWithoutMessage", func() {
		// Arrange
		update := &models.Update{Message: nil}

		// Act - should not panic and should not publish events
		handler := suite.handlers.UrlHandler()
		handler(suite.ctx, suite.bot, update)

		// Assert - no events should be published
	})

	suite.Run("IgnoresUpdateWithEmptyText", func() {
		// Arrange
		message := suite.createMessage("", 12345, 67890, 333)
		update := suite.createUpdate(message)

		// Act - should not panic and should not publish events
		handler := suite.handlers.UrlHandler()
		handler(suite.ctx, suite.bot, update)

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

		suite.bot.EXPECT().SendMessage(suite.ctx, params).Return(expectedMessage, nil)

		// Act
		suite.handlers.sendMessage(suite.ctx, suite.bot, params)

		// Assert
		suite.bot.AssertExpectations(suite.T())
	})

	suite.Run("HandlesErrorGracefully", func() {
		// Arrange
		params := &bot.SendMessageParams{
			ChatID: 67890,
			Text:   "test message",
		}

		suite.bot.EXPECT().SendMessage(suite.ctx, params).Return(nil, domain.ErrDownloadFailed)

		// Act - should not panic
		suite.NotPanics(func() {
			suite.handlers.sendMessage(suite.ctx, suite.bot, params)
		})

		// Assert
		suite.bot.AssertExpectations(suite.T())
	})
}

// Test AllowedUsersMiddleware
func (suite *TestTelegramHandlersSuite) TestAllowedUsersMiddleware() {
	suite.Run("AllowsUpdateFromAllowedMessageUser", func() {
		update := suite.createUpdate(&models.Message{
			ID: 1,
			From: &models.User{
				ID: 111,
			},
			Chat: models.Chat{ID: 2},
			Text: "hello",
		})

		allowed := []int64{111}
		called := false
		mw := suite.handlers.AllowedUsersMiddleware(allowed)
		next := func(ctx context.Context, b telegram.Bot, update *models.Update) {
			called = true
		}

		handler := mw(next)
		handler(suite.ctx, suite.bot, update)
		suite.True(called, "next should be called for allowed message user")
	})

	suite.Run("AllowsUpdateFromAllowedCallbackQueryUser", func() {
		update := &models.Update{
			CallbackQuery: &models.CallbackQuery{
				From: models.User{ID: 222},
			},
		}

		allowed := []int64{222}
		called := false
		mw := suite.handlers.AllowedUsersMiddleware(allowed)
		next := func(ctx context.Context, b telegram.Bot, update *models.Update) {
			called = true
		}

		handler := mw(next)
		handler(suite.ctx, suite.bot, update)
		suite.True(called, "next should be called for allowed callback query user")
	})

	suite.Run("BlocksUpdateFromNotAllowedUser", func() {
		update := suite.createUpdate(&models.Message{
			ID:   2,
			From: &models.User{ID: 333},
			Chat: models.Chat{ID: 3},
			Text: "blocked",
		})

		allowed := []int64{444}
		called := false
		mw := suite.handlers.AllowedUsersMiddleware(allowed)
		next := func(ctx context.Context, b telegram.Bot, update *models.Update) {
			called = true
		}

		handler := mw(next)
		handler(suite.ctx, suite.bot, update)
		suite.False(called, "next should NOT be called for disallowed user")
	})

	suite.Run("BlocksUpdateWithUnknownUser", func() {
		// update without message, callbackQuery, inlineQuery or editedMessage
		update := &models.Update{}

		allowed := []int64{111}
		called := false
		mw := suite.handlers.AllowedUsersMiddleware(allowed)
		next := func(ctx context.Context, b telegram.Bot, update *models.Update) {
			called = true
		}

		handler := mw(next)
		handler(suite.ctx, suite.bot, update)
		suite.False(called, "next should NOT be called when user ID cannot be determined")
	})
}

// Test CmdInfoHandler method

func (suite *TestTelegramHandlersSuite) TestHandleInfo() {
	suite.Run("PublishesFeedInfoRequestEvent", func() {
		// Arrange
		update := suite.createUpdate(&models.Message{
			ID:   10,
			From: &models.User{ID: 555},
			Chat: models.Chat{ID: 66},
			Text: "/info",
		})

		expectedSource := domain.RequestSource{
			UserID:    update.Message.From.ID,
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
		}

		suite.bus.EXPECT().Publish(suite.matchFeedInfoRequestEvent(expectedSource)).Return()

		// Act
		handler := suite.handlers.CmdInfoHandler()
		handler(suite.ctx, suite.bot, update)

		// Assert
		suite.bus.AssertExpectations(suite.T())
	})

	suite.Run("IgnoresUpdateWithoutMessage", func() {
		// Arrange
		update := &models.Update{Message: nil}

		// Act - should not panic and should not publish events
		handler := suite.handlers.CmdInfoHandler()
		handler(suite.ctx, suite.bot, update)

		// No EXPECT on bus, so nothing to assert beyond no panic
	})
}

// Helper matchers for mock expectations

func (suite *TestTelegramHandlersSuite) matchSendMessageParams(expected *bot.SendMessageParams) interface{} {
	return mock.MatchedBy(func(params *bot.SendMessageParams) bool {
		return params.ChatID == expected.ChatID &&
			params.Text == expected.Text &&
			params.ParseMode == expected.ParseMode
	})
}

func (suite *TestTelegramHandlersSuite) matchBuildRequestEvent(expectedSource domain.RequestSource) interface{} {
	return mock.MatchedBy(func(event domain.Event) bool {
		if event.Type() != domain.BuildRequestEvent {
			return false
		}

		req, ok := event.Payload().(domain.BuildRequest)
		if !ok {
			return false
		}

		if req.Source == nil {
			return false
		}

		return req.Source.UserID == expectedSource.UserID &&
			req.Source.ChatID == expectedSource.ChatID &&
			req.Source.MessageID == expectedSource.MessageID
	})
}

func (suite *TestTelegramHandlersSuite) matchDownloadRequestEvent(expectedSource domain.RequestSource, expectedUrl string) interface{} {
	return mock.MatchedBy(func(event domain.Event) bool {
		if event.Type() != domain.DownloadRequestEvent {
			return false
		}

		req, ok := event.Payload().(domain.DownloadRequest)
		if !ok {
			return false
		}

		return req.Source.UserID == expectedSource.UserID &&
			req.Source.ChatID == expectedSource.ChatID &&
			req.Source.MessageID == expectedSource.MessageID &&
			req.Url == expectedUrl
	})
}

func (suite *TestTelegramHandlersSuite) matchFeedInfoRequestEvent(expectedSource domain.RequestSource) interface{} {
	return mock.MatchedBy(func(event domain.Event) bool {
		if event.Type() != domain.FeedInfoRequestEvent {
			return false
		}
		req, ok := event.Payload().(domain.FeedInfoRequest)
		if !ok {
			return false
		}
		return req.Source.UserID == expectedSource.UserID &&
			req.Source.ChatID == expectedSource.ChatID &&
			req.Source.MessageID == expectedSource.MessageID
	})
}
