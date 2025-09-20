package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/internal/config"
)

// Middleware represents a Telegram bot middleware.
type Middleware struct {
	cfg config.Telegram
	log *slog.Logger
}

// NewMiddleware creates a new Middleware instance.
func NewMiddleware(cfg config.Telegram, log *slog.Logger) *Middleware {
	return &Middleware{
		cfg: cfg,
		log: log,
	}
}

// WithAllowedUsers is a middleware that blocks updates from users not in the allowed users list.
func (m *Middleware) WithAllowedUsers() bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			var userID int64

			// Extract user ID from the update
			if update.Message != nil && update.Message.From != nil {
				userID = update.Message.From.ID
			} else if update.CallbackQuery != nil {
				userID = update.CallbackQuery.From.ID
			} else if update.InlineQuery != nil {
				userID = update.InlineQuery.From.ID
			} else if update.EditedMessage != nil && update.EditedMessage.From != nil {
				userID = update.EditedMessage.From.ID
			} else {
				// If user ID cannot be determined, block the update
				m.log.Error("[bot] update blocked: cannot determine user ID", "update_id", update.ID)
				return
			}

			// Check if user is allowed
			allowed := false
			for _, allowedUserID := range m.cfg.AllowedUsers {
				if userID == allowedUserID {
					allowed = true
					break
				}
			}

			if !allowed {
				m.log.Error("[bot] update blocked: user not allowed",
					"user_id", userID, "update_id", update.ID)
				return
			}

			next(ctx, b, update)
		}
	}
}
