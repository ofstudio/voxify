package middleware

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/pkg/telegram"
)

// AllowedUsers is a Telegram middleware that blocks updates from users not in the allowed users list.
func AllowedUsers(log *slog.Logger, allowedUsers []int64) telegram.Middleware {
	return func(next telegram.HandlerFunc) telegram.HandlerFunc {
		return func(ctx context.Context, api telegram.API, update *models.Update) {
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
				log.Error("[telegram handlers] update blocked: cannot determine user ID",
					"update", telegram.LogUpdate(update))
				return
			}

			// Check if user is allowed
			allowed := false
			for _, allowedUserID := range allowedUsers {
				if userID == allowedUserID {
					allowed = true
					break
				}
			}

			if !allowed {
				log.Error("[telegram handlers] update blocked: user not allowed",
					"update", telegram.LogUpdate(update))
				return
			}

			next(ctx, api, update)
		}
	}
}
