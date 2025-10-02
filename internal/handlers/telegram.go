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

// Start initializes handlers and starts the Telegram bot.
func (h *TelegramHandlers) Start(ctx context.Context) error {
	h.log.Info("[telegram handlers] starting bot")

	// Register handlers
	h.bot.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommandStartOnly, h.cmdStartHandler())
	h.bot.RegisterHandler(bot.HandlerTypeMessageText, "build", bot.MatchTypeCommand, h.cmdBuildHandler())
	h.bot.RegisterHandler(bot.HandlerTypeMessageText, "info", bot.MatchTypeCommand, h.cmdInfoHandler())
	h.bot.RegisterHandler(bot.HandlerTypeMessageText, "https://", bot.MatchTypePrefix, h.urlHandler())

	//start the bot
	go func() {
		h.bot.Start(ctx)
		h.log.Info("[telegram handlers] bot stopped")
	}()

	h.log.Info("[telegram handlers] started")
	return nil
}

// ErrorsHandler handles errors from the Telegram bot.
func (h *TelegramHandlers) ErrorsHandler() bot.ErrorsHandler {
	return func(err error) {
		h.log.Error("[telegram handlers] telegram error", slog.String("error", err.Error()))
	}
}

// cmdStartHandler handles the /start command.
func (h *TelegramHandlers) cmdStartHandler() telegram.HandlerFunc {
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

// cmdBuildHandler handles the /build command.
func (h *TelegramHandlers) cmdBuildHandler() telegram.HandlerFunc {
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

// cmdInfoHandler handles the /info command.
func (h *TelegramHandlers) cmdInfoHandler() telegram.HandlerFunc {
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

// urlHandler handles messages containing URLs.
func (h *TelegramHandlers) urlHandler() telegram.HandlerFunc {
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
