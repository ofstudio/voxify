package handlers

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/templates"
	"github.com/ofstudio/voxify/pkg/randtoken"
	"github.com/ofstudio/voxify/pkg/telegram"
)

// TelegramHandlers handles Telegram bot updates and commands.
type TelegramHandlers struct {
	log *slog.Logger
	bus domain.EventBus
	bot telegram.Bot
}

// NewTelegramHandlers creates a new instance of TelegramHandlers.
func NewTelegramHandlers(log *slog.Logger, bus domain.EventBus, bot telegram.Bot) *TelegramHandlers {
	return &TelegramHandlers{log: log, bus: bus, bot: bot}
}

// ErrorsHandler handles errors from the Telegram bot.
func (h *TelegramHandlers) ErrorsHandler() bot.ErrorsHandler {
	return func(err error) {
		h.log.Error("[telegram handlers] telegram error", slog.String("error", err.Error()))
	}
}

// CmdStartHandler handles the /start command.
func (h *TelegramHandlers) CmdStartHandler() telegram.HandlerFunc {
	return func(ctx context.Context, api telegram.API, update *models.Update) {
		if update.Message == nil {
			return
		}
		h.log.Info("[bot] start command received",
			"update", telegram.LogUpdate(update))

		// Send welcome message
		h.sendMessage(ctx, api, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      templates.MsgStart,
			ParseMode: models.ParseModeHTML,
		})
	}
}

// CmdBuildHandler handles the /build command.
func (h *TelegramHandlers) CmdBuildHandler() telegram.HandlerFunc {
	return func(ctx context.Context, api telegram.API, update *models.Update) {
		if update.Message == nil {
			return
		}
		h.log.Info("[bot] build command received",
			"update", telegram.LogUpdate(update))

		// Publish build request
		src := h.requestSource(update.Message)
		h.bus.Publish(domain.NewBuildRequestEvent(domain.BuildRequest{
			ID:     h.requestID(),
			Source: &src,
		}))
	}
}

// CmdInfoHandler handles the /info command.
func (h *TelegramHandlers) CmdInfoHandler() telegram.HandlerFunc {
	return func(ctx context.Context, api telegram.API, update *models.Update) {
		if update.Message == nil {
			return
		}
		h.log.Info("[bot] info command received",
			"update", telegram.LogUpdate(update))

		// Publish info request
		h.bus.Publish(domain.NewFeedInfoRequestEvent(domain.FeedInfoRequest{
			ID:     h.requestID(),
			Source: h.requestSource(update.Message),
		}))
	}
}

// UrlHandler handles messages containing URLs.
func (h *TelegramHandlers) UrlHandler() telegram.HandlerFunc {
	return func(ctx context.Context, api telegram.API, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		h.log.Info("[bot] url message received",
			"update", telegram.LogUpdate(update), "url", update.Message.Text)

		// Publish download request
		h.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
			ID:     h.requestID(),
			Source: h.requestSource(update.Message),
			Url:    update.Message.Text,
		}))
	}
}

// AllowedUsersMiddleware is a middleware that blocks updates from users not in the allowed users list.
func (h *TelegramHandlers) AllowedUsersMiddleware(allowedUsers []int64) telegram.Middleware {
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
				h.log.Error("[telegram handlers] update blocked: cannot determine user ID",
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
				h.log.Error("[telegram handlers] update blocked: user not allowed",
					"update", telegram.LogUpdate(update))
				return
			}

			next(ctx, api, update)
		}
	}
}

// requestSource creates a RequestSource from a Telegram message.
func (h *TelegramHandlers) requestSource(msg *models.Message) domain.RequestSource {
	return domain.RequestSource{
		UserID:    msg.From.ID,
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	}
}

func (h *TelegramHandlers) requestID() string {
	return randtoken.New(10)
}

func (h *TelegramHandlers) sendMessage(ctx context.Context, api telegram.API, p *bot.SendMessageParams) {
	msg, err := api.SendMessage(ctx, p)
	if err != nil {
		h.log.Error("[telegram handlers] failed to send message",
			"error", err.Error(), "chat_id", p.ChatID)
		return
	}

	h.log.Info("[telegram handlers] message sent",
		"message", telegram.LogMessage(msg),
	)
}
