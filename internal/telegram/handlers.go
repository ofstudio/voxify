package telegram

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/locales"
)

type Handlers struct {
	log       *slog.Logger
	processor Processor
	builder   Builder
	settings  config.Settings
}

func NewHandlers(settings config.Settings, log *slog.Logger, processor Processor, builder Builder) *Handlers {
	return &Handlers{
		log:       log,
		processor: processor,
		builder:   builder,
		settings:  settings,
	}
}

func (h *Handlers) Error() bot.ErrorsHandler {
	return func(err error) {
		h.log.Error("[bot] telegram error", slog.String("error", err.Error()))
	}
}

func (h *Handlers) CmdStart() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		h.log.Info("[bot] start command received", "update_id", update.ID, "message", logMessage(update.Message))

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   locales.MsgStart,
		})

		if err != nil {
			h.log.Error("[bot] failed to send start message",
				"error", err.Error(), "chat", logChat(&update.Message.Chat))
		} else {
			h.log.Info("[bot] start message sent", "chat", logChat(&update.Message.Chat))
		}
	}
}

func (h *Handlers) CmdBuild() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		h.log.Info("[bot] build command received", "update_id", update.ID, "message", logMessage(update.Message))
		err := h.builder.Build(ctx)
		if err != nil {
			h.log.Error("[bot] failed to build podcast feed",
				"error", err.Error(), "chat", logChat(&update.Message.Chat))
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   locales.MsgBuildError,
			})
			return
		}

		if _, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   locales.MsgBuildSuccess,
		}); err != nil {
			h.log.Error("[bot] failed to send build success message",
				"error", err.Error(), "chat", logChat(&update.Message.Chat))
		}
	}
}

func (h *Handlers) Url() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		h.log.Info("[bot] url received", "update_id", update.ID, "url", update.Message.Text)
		request := &entities.Request{
			UserID:    update.Message.From.ID,
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
			Url:       update.Message.Text,
			Force:     false,
		}

		h.processor.In() <- request
	}
}
