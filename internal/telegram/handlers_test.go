package telegram

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/mocks"
)

// TestHandlersSuite is a test suite for Handlers
type TestHandlersSuite struct {
	suite.Suite
	ctx         context.Context
	log         *slog.Logger
	cfg         config.Settings
	handlers    *Handlers
	requestChan chan entities.Request
	mockFeeder  *mocks.MockFeeder
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestHandlersSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	suite.cfg = config.Settings{
		// Add any required config settings here
	}
}

// SetupTest is called before each test method
func (suite *TestHandlersSuite) SetupTest() {
	suite.requestChan = make(chan entities.Request, 10)
	suite.mockFeeder = mocks.NewMockFeeder(suite.T())
	suite.handlers = NewHandlers(suite.cfg, suite.log, suite.requestChan, suite.mockFeeder)
}

// TestSendRequest tests the sendRequest method
func (suite *TestHandlersSuite) TestSendRequest() {
	suite.Run("Success", func() {
		// Arrange
		request := entities.Request{
			UserID:    123,
			ChatID:    456,
			MessageID: 789,
			Url:       "https://example.com/video",
			Force:     false,
		}

		// Act
		err := suite.handlers.sendRequest(suite.ctx, request)

		// Assert
		suite.NoError(err)

		// Verify the request was sent to the channel
		select {
		case receivedReq := <-suite.requestChan:
			suite.Equal(request, receivedReq)
		case <-time.After(100 * time.Millisecond):
			suite.Fail("Request was not sent to channel")
		}
	})

	suite.Run("ProcessorBusy", func() {
		// Arrange - fill the channel to simulate busy processor
		for i := 0; i < cap(suite.requestChan); i++ {
			suite.requestChan <- entities.Request{}
		}

		request := entities.Request{
			UserID:    123,
			ChatID:    456,
			MessageID: 789,
			Url:       "https://example.com/video",
			Force:     false,
		}

		// Act
		err := suite.handlers.sendRequest(suite.ctx, request)

		// Assert
		suite.Error(err)
		suite.True(errors.Is(err, errProcessorBusy))
	})

	suite.Run("ContextCanceled", func() {
		// Arrange
		cancelCtx, cancel := context.WithCancel(suite.ctx)
		cancel() // Cancel the context immediately

		request := entities.Request{
			UserID:    123,
			ChatID:    456,
			MessageID: 789,
			Url:       "https://example.com/video",
			Force:     false,
		}

		// Act
		err := suite.handlers.sendRequest(cancelCtx, request)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "context done")
	})
}

// TestError tests the Error handler
func (suite *TestHandlersSuite) TestError() {
	suite.Run("HandlesError", func() {
		// Arrange
		errorHandler := suite.handlers.Error()
		testError := errors.New("test telegram error")

		// Act - this should not panic
		errorHandler(testError)

		// Assert - if we get here, the error was handled without panic
		suite.True(true)
	})
}

// TestGetInfoMessage tests the getInfoMessage method
func (suite *TestHandlersSuite) TestGetInfoMessage() {
	suite.Run("Success", func() {
		// Arrange: use isolated mock and handlers
		mockFeeder := mocks.NewMockFeeder(suite.T())
		reqCh := make(chan entities.Request, 1)
		h := NewHandlers(suite.cfg, suite.log, reqCh, mockFeeder)

		feed := &entities.Feed{
			Title:       "My podcast",
			Description: "Awesome show",
			Language:    "en",
			Categories: []entities.FeedCategory{
				{Text: "Science", Subcategories: []string{"Physics", "Astronomy"}},
				{Text: "Technology"},
			},
			Keywords:     "Tech,Talks",
			Author:       "Alice",
			Explicit:     true,
			WebsiteLink:  "https://site.example",
			ImageUrl:     "https://site.example/cover.jpg",
			EpisodeCount: 3,
			RSSLink:      "https://site.example/rss.xml",
		}
		mockFeeder.On("Feed", suite.ctx).Return(feed, nil).Once()

		// Act
		msg, err := h.getInfoMessage(suite.ctx)

		// Assert
		suite.NoError(err)
		suite.NotEmpty(msg)
		suite.Contains(msg, "My podcast")
		suite.Contains(msg, "Awesome show")
		suite.Contains(msg, "By Alice")
		suite.Contains(msg, "Language: en")
		suite.Contains(msg, "Categories: Science, Physics, Astronomy, Technology")
		suite.Contains(msg, "Keywords: Tech,Talks")
		suite.Contains(msg, "<a href=\"https://site.example/cover.jpg\">Artwork</a>")
		suite.Contains(msg, "<a href=\"https://site.example\">Website</a>")
		suite.Contains(msg, "Number of episodes: 3")
		suite.Contains(msg, "Explicit content")
		suite.Contains(msg, "RSS: https://site.example/rss.xml")
	})

	suite.Run("Error", func() {
		// Arrange: use isolated mock and handlers
		mockFeeder := mocks.NewMockFeeder(suite.T())
		reqCh := make(chan entities.Request, 1)
		h := NewHandlers(suite.cfg, suite.log, reqCh, mockFeeder)

		expectedErr := errors.New("feed error")
		mockFeeder.On("Feed", suite.ctx).Return((*entities.Feed)(nil), expectedErr).Once()

		// Act
		msg, err := h.getInfoMessage(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Equal("", msg)
		suite.Contains(err.Error(), "failed to get feed info")
	})
}

// TestCategoriesToString tests the categoriesToString helper
func (suite *TestHandlersSuite) TestCategoriesToString() {
	suite.Run("WithSubcategories", func() {
		// Arrange
		cats := []entities.FeedCategory{
			{Text: "Science", Subcategories: []string{"Physics", "Astronomy"}},
			{Text: "Technology"},
		}
		// Act
		res := categoriesToString(cats)
		// Assert
		suite.Equal("Science, Physics, Astronomy, Technology", res)
	})

	suite.Run("Empty", func() {
		// Arrange
		var cats []entities.FeedCategory
		// Act
		res := categoriesToString(cats)
		// Assert
		suite.Equal("", res)
	})
}

// Run the test suite
func TestHandlers(t *testing.T) {
	suite.Run(t, new(TestHandlersSuite))
}
