package middleware

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/pkg/telegram"
	"github.com/stretchr/testify/suite"
)

func TestAllowedUserTestSuite(t *testing.T) {
	suite.Run(t, new(AllowedUsersTestSuite))
}

type AllowedUsersTestSuite struct {
	suite.Suite
	ctx    context.Context
	logger *slog.Logger
}

func (suite *AllowedUsersTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func (suite *AllowedUsersTestSuite) TestAllowedUsers() {
	suite.Run("Message", func() {
		update := &models.Update{
			Message: &models.Message{
				ID: 1,
				From: &models.User{
					ID: 111,
				},
				Chat: models.Chat{ID: 2},
				Text: "hello",
			},
		}

		allowedUsers := []int64{111}
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.True(called, "next should be called for allowed message user")
	})

	suite.Run("CallbackQuery", func() {
		update := &models.Update{
			CallbackQuery: &models.CallbackQuery{
				From: models.User{ID: 222},
			},
		}

		allowedUsers := []int64{222}
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.True(called, "next should be called for allowed callback query user")
	})

	suite.Run("InlineQuery", func() {
		update := &models.Update{
			InlineQuery: &models.InlineQuery{
				From: &models.User{ID: 333},
			},
		}

		allowedUsers := []int64{333}
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.True(called, "next should be called for allowed inline query user")
	})

	suite.Run("EditedMessage", func() {
		update := &models.Update{
			EditedMessage: &models.Message{
				ID: 4,
				From: &models.User{
					ID: 444,
				},
				Chat: models.Chat{ID: 5},
				Text: "edited message",
			},
		}

		allowedUsers := []int64{444}
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.True(called, "next should be called for allowed edited message user")
	})

	suite.Run("MultipleUsersAllowed", func() {
		allowedUsers := []int64{111, 222, 333}

		// Test first user
		update1 := &models.Update{
			Message: &models.Message{
				From: &models.User{ID: 111},
				Chat: models.Chat{ID: 1},
			},
		}

		called1 := false
		middleware := AllowedUsers(suite.logger, allowedUsers)
		next1 := func(ctx context.Context, api telegram.API, update *models.Update) {
			called1 = true
		}

		handler1 := middleware(next1)
		handler1(suite.ctx, nil, update1)
		suite.True(called1, "first user should be allowed")

		// Test second user
		update2 := &models.Update{
			CallbackQuery: &models.CallbackQuery{
				From: models.User{ID: 222},
			},
		}

		called2 := false
		next2 := func(ctx context.Context, api telegram.API, update *models.Update) {
			called2 = true
		}

		handler2 := middleware(next2)
		handler2(suite.ctx, nil, update2)
		suite.True(called2, "second user should be allowed")

		// Test third user
		update3 := &models.Update{
			InlineQuery: &models.InlineQuery{
				From: &models.User{ID: 333},
			},
		}

		called3 := false
		next3 := func(ctx context.Context, api telegram.API, update *models.Update) {
			called3 = true
		}

		handler3 := middleware(next3)
		handler3(suite.ctx, nil, update3)
		suite.True(called3, "third user should be allowed")
	})
}

func (suite *AllowedUsersTestSuite) TestBlockUsers() {
	suite.Run("NotAllowedUser", func() {
		update := &models.Update{
			Message: &models.Message{
				ID:   6,
				From: &models.User{ID: 555},
				Chat: models.Chat{ID: 7},
				Text: "blocked message",
			},
		}

		allowedUsers := []int64{666} // Different user ID
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.False(called, "next should not be called for not allowed user")
	})

	suite.Run("WithoutUserID", func() {
		update := &models.Update{
			// No Message, CallbackQuery, InlineQuery, or EditedMessage
		}

		allowedUsers := []int64{111}
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.False(called, "next should not be called when user ID cannot be determined")
	})

	suite.Run("MessageWithoutFrom", func() {
		update := &models.Update{
			Message: &models.Message{
				ID:   8,
				From: nil, // No From field
				Chat: models.Chat{ID: 9},
				Text: "message without from",
			},
		}

		allowedUsers := []int64{111}
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.False(called, "next should not be called when message has no from field")
	})

	suite.Run("EditedMessageWithoutFrom", func() {
		update := &models.Update{
			EditedMessage: &models.Message{
				ID:   10,
				From: nil, // No From field
				Chat: models.Chat{ID: 11},
				Text: "edited message without from",
			},
		}

		allowedUsers := []int64{111}
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.False(called, "next should not be called when edited message has no from field")
	})

	suite.Run("EmptyAllowedList", func() {
		update := &models.Update{
			Message: &models.Message{
				From: &models.User{ID: 111},
				Chat: models.Chat{ID: 1},
			},
		}

		allowedUsers := []int64{} // Empty list
		called := false

		middleware := AllowedUsers(suite.logger, allowedUsers)
		next := func(ctx context.Context, api telegram.API, update *models.Update) {
			called = true
		}

		handler := middleware(next)
		handler(suite.ctx, nil, update)

		suite.False(called, "next should not be called when allowed users list is empty")
	})
}
