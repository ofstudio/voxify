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
	mockBuilder *mocks.MockBuilder
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
	suite.mockBuilder = mocks.NewMockBuilder(suite.T())
	suite.handlers = NewHandlers(suite.cfg, suite.log, suite.requestChan, suite.mockBuilder)
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

// Run the test suite
func TestHandlers(t *testing.T) {
	suite.Run(t, new(TestHandlersSuite))
}
