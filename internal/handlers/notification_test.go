package handlers

import (
	"context"
	"errors"
	"html/template"
	"log/slog"
	"strings"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/mocks"
	"github.com/ofstudio/voxify/internal/templates"
)

// TestNotificationHandlers is the entry point for running the test suite
func TestNotificationHandlers(t *testing.T) {
	suite.Run(t, new(TestNotificationHandlersSuite))
}

// TestNotificationHandlersSuite is a test suite for NotificationHandlers
type TestNotificationHandlersSuite struct {
	suite.Suite
	ctx      context.Context
	cancel   context.CancelFunc
	log      *slog.Logger
	bus      *mocks.MockEventBus
	api      *mocks.MockAPI
	handlers *NotificationHandlers
}

func (suite *TestNotificationHandlersSuite) SetupTest() {
	suite.ctx, suite.cancel = context.WithCancel(context.Background())
	suite.log = slog.Default()
	suite.bus = mocks.NewMockEventBus(suite.T())
	suite.api = mocks.NewMockAPI(suite.T())

	suite.handlers = NewNotificationHandlers(suite.log, suite.bus).WithAPI(suite.api)
	// Manually set context because in these tests we call handler methods directly (without Init)
	suite.handlers.ctx = suite.ctx
}

func (suite *TestNotificationHandlersSuite) TearDownTest() {
	if suite.cancel != nil {
		suite.cancel()
	}
}

// Helpers
func (suite *TestNotificationHandlersSuite) matchSendMessageParams(expectedChatID int64, expectedTextContains string) interface{} {
	return mock.MatchedBy(func(p *bot.SendMessageParams) bool {
		if p == nil || p.ChatID != expectedChatID {
			return false
		}
		if expectedTextContains == "" {
			return true
		}
		return strings.Contains(p.Text, expectedTextContains)
	})
}

// Tests
func (suite *TestNotificationHandlersSuite) TestNewNotificationHandlers() {
	h := NewNotificationHandlers(suite.log, suite.bus)
	suite.NotNil(h)
	suite.Equal(suite.log, h.log)
	suite.Equal(suite.bus, h.bus)
	suite.Nil(h.api) // api should be nil until WithAPI is called
}

func (suite *TestNotificationHandlersSuite) TestWithAPI() {
	// Arrange
	handlers := NewNotificationHandlers(suite.log, suite.bus)

	// Act
	result := handlers.WithAPI(suite.api)

	// Assert
	suite.Equal(handlers, result) // should return same instance for chaining
	suite.Equal(suite.api, handlers.api)
}

func (suite *TestNotificationHandlersSuite) TestStart() {
	suite.Run("ReturnsErrorWhenAPINotSet", func() {
		// Arrange
		handlers := NewNotificationHandlers(suite.log, suite.bus)

		// Act
		err := handlers.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "telegram API is not set")
	})

	suite.Run("ReturnsErrorWhenEventBusNotSet", func() {
		// Arrange
		handlers := NewNotificationHandlers(suite.log, nil).WithAPI(suite.api)

		// Act
		err := handlers.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "event bus is not set")
	})

	suite.Run("SubscribesHandlersSuccessfully", func() {
		// Arrange
		// Expect subscriptions for three response event types
		suite.bus.EXPECT().Subscribe(domain.DownloadResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.BuildResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
		suite.bus.EXPECT().Subscribe(domain.FeedInfoResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Act
		err := suite.handlers.Init(ctx)

		// Assert
		suite.NoError(err)
		suite.bus.AssertExpectations(suite.T())
	})
}

func (suite *TestNotificationHandlersSuite) TestStartSubscribesHandlers() {
	// Expect subscriptions for three response event types
	suite.bus.EXPECT().Subscribe(domain.DownloadResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
	suite.bus.EXPECT().Subscribe(domain.BuildResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()
	suite.bus.EXPECT().Subscribe(domain.FeedInfoResponseEvent, mock.AnythingOfType("domain.EventHandler")).Return()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	suite.NoError(suite.handlers.Init(ctx))

	// Assert
	suite.bus.AssertExpectations(suite.T())
}

func (suite *TestNotificationHandlersSuite) TestDownloadNotificationHandler() {
	src := domain.RequestSource{UserID: 1, ChatID: 10, MessageID: 100}

	suite.Run("Pending_ShouldSendStartedMessage", func() {
		resp := domain.DownloadResponse{Status: domain.StatusPending, Request: domain.DownloadRequest{Source: src}}
		event := domain.NewDownloadResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgDownloadStarted)).Return(&models.Message{}, nil)
		suite.handlers.downloadNotificationHandler(event)
	})

	suite.Run("Success_ShouldSendSuccessMessageWithTitle", func() {
		ep := &domain.Episode{Title: "Awesome Episode"}
		resp := domain.DownloadResponse{Status: domain.StatusSuccess, Episode: ep, Request: domain.DownloadRequest{Source: src}}
		event := domain.NewDownloadResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, ep.Title)).Return(&models.Message{}, nil)
		suite.handlers.downloadNotificationHandler(event)
	})

	suite.Run("Failed_ShouldSendErrorMessage", func() {
		testErr := errors.New("download failed")
		resp := domain.DownloadResponse{Status: domain.StatusFailed, Error: testErr, Request: domain.DownloadRequest{Source: src}}
		event := domain.NewDownloadResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgError(testErr))).Return(&models.Message{}, nil)
		suite.handlers.downloadNotificationHandler(event)
	})

	suite.Run("UnknownStatus_ShouldSendSomethingWentWrong", func() {
		resp := domain.DownloadResponse{Status: domain.ResponseStatus(999), Request: domain.DownloadRequest{Source: src}}
		event := domain.NewDownloadResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgSomethingWentWrong)).Return(&models.Message{}, nil)
		suite.handlers.downloadNotificationHandler(event)
	})
}

func (suite *TestNotificationHandlersSuite) TestBuildNotificationHandler() {
	src := domain.RequestSource{UserID: 2, ChatID: 20, MessageID: 200}

	suite.Run("NilSource_ShouldNotSend", func() {
		resp := domain.BuildResponse{Status: domain.StatusSuccess, Request: domain.BuildRequest{Source: nil}}
		event := domain.NewBuildResponseEvent(resp)
		// No expectation
		suite.handlers.buildNotificationHandler(event)
	})

	suite.Run("Success_ShouldSendBuildSuccess", func() {
		req := domain.BuildRequest{Source: &src}
		resp := domain.BuildResponse{Status: domain.StatusSuccess, Request: req}
		event := domain.NewBuildResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgBuildSuccess)).Return(&models.Message{}, nil)
		suite.handlers.buildNotificationHandler(event)
	})

	suite.Run("Failed_ShouldSendErrorMessage", func() {
		req := domain.BuildRequest{Source: &src}
		testErr := errors.New("build failed")
		resp := domain.BuildResponse{Status: domain.StatusFailed, Error: testErr, Request: req}
		event := domain.NewBuildResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgError(testErr))).Return(&models.Message{}, nil)
		suite.handlers.buildNotificationHandler(event)
	})

	suite.Run("Pending_ShouldNoOp", func() {
		req := domain.BuildRequest{Source: &src}
		resp := domain.BuildResponse{Status: domain.StatusPending, Request: req}
		event := domain.NewBuildResponseEvent(resp)
		suite.handlers.buildNotificationHandler(event)
	})

	suite.Run("UnknownStatus_ShouldSendSomethingWentWrong", func() {
		req := domain.BuildRequest{Source: &src}
		resp := domain.BuildResponse{Status: domain.ResponseStatus(777), Request: req}
		event := domain.NewBuildResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgSomethingWentWrong)).Return(&models.Message{}, nil)
		suite.handlers.buildNotificationHandler(event)
	})
}

func (suite *TestNotificationHandlersSuite) TestFeedInfoNotificationHandler() {
	src := domain.RequestSource{UserID: 3, ChatID: 30, MessageID: 300}

	suite.Run("Success_ShouldRenderTemplateAndSend", func() {
		// Provide a minimal template to avoid nil pointer (in case init failed)
		templates.FeedInfoTemplate = template.Must(template.New("feed_info").Parse("<b>{{.Title}}</b> {{.Description}}"))
		fi := &domain.FeedInfo{Title: "My Show", Description: "Desc"}
		resp := domain.FeedInfoResponse{Status: domain.StatusSuccess, FeedInfo: fi, Request: domain.FeedInfoRequest{Source: src}}
		event := domain.NewFeedInfoResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, fi.Title)).Return(&models.Message{}, nil)
		suite.handlers.feedInfoNotificationHandler(event)
	})

	suite.Run("Failed_ShouldSendErrorMessage", func() {
		testErr := errors.New("feed failed")
		resp := domain.FeedInfoResponse{Status: domain.StatusFailed, Error: testErr, Request: domain.FeedInfoRequest{Source: src}}
		event := domain.NewFeedInfoResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgError(testErr))).Return(&models.Message{}, nil)
		suite.handlers.feedInfoNotificationHandler(event)
	})

	suite.Run("Pending_ShouldNoOp", func() {
		resp := domain.FeedInfoResponse{Status: domain.StatusPending, Request: domain.FeedInfoRequest{Source: src}}
		event := domain.NewFeedInfoResponseEvent(resp)
		suite.handlers.feedInfoNotificationHandler(event)
	})

	suite.Run("UnknownStatus_ShouldSendSomethingWentWrong", func() {
		resp := domain.FeedInfoResponse{Status: domain.ResponseStatus(555), Request: domain.FeedInfoRequest{Source: src}}
		event := domain.NewFeedInfoResponseEvent(resp)

		suite.api.EXPECT().SendMessage(suite.ctx, suite.matchSendMessageParams(src.ChatID, templates.MsgSomethingWentWrong)).Return(&models.Message{}, nil)
		suite.handlers.feedInfoNotificationHandler(event)
	})
}
