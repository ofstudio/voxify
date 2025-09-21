package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/locales"
)

type Handlers struct {
	log    *slog.Logger
	cfg    config.Settings
	out    chan<- entities.Request
	feeder Feeder
}

func NewHandlers(cfg config.Settings, log *slog.Logger, out chan<- entities.Request, f Feeder) *Handlers {
	return &Handlers{
		log:    log,
		cfg:    cfg,
		feeder: f,
		out:    out,
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
		h.sendMessage(ctx, b, update.Message.Chat, locales.MsgStart)
	}
}

func (h *Handlers) CmdBuild() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		h.log.Info("[bot] build command received", "update_id", update.ID, "message", logMessage(update.Message))

		msg := locales.MsgBuildSuccess
		if err := h.feeder.Build(ctx); err != nil {
			msg = msgErr(err)
			h.log.Error("[bot] failed to build podcast feed",
				"error", err.Error(), "chat", logChat(&update.Message.Chat))
		}
		h.sendMessage(ctx, b, update.Message.Chat, msg)
	}
}

func (h *Handlers) Url() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.Text == "" {
			return
		}
		h.log.Info("[bot] url received", "update_id", update.ID, "url", update.Message.Text)
		request := entities.Request{
			UserID:    update.Message.From.ID,
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.ID,
			Url:       update.Message.Text,
			Force:     false,
		}
		if err := h.sendRequest(ctx, request); err != nil {
			h.log.Error("[bot] failed to queue request",
				"error", err.Error(), "request", request.LogValue())
			h.sendMessage(ctx, b, update.Message.Chat, msgErr(err))
		}
	}
}

// sendRequest tries to send the request to the processor safely.
func (h *Handlers) sendRequest(ctx context.Context, req entities.Request) error {
	select {
	// if the request was successfully sent to the processor
	case h.out <- req:
		return nil
	// if the processor is busy (no worker available in 2 seconds)
	case <-time.After(time.Second * 2):
		return errProcessorBusy
	// if the context is done (app shutting down)
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	}
}

// sendMessage sends text message to the specified chat.
func (h *Handlers) sendMessage(ctx context.Context, b *bot.Bot, c models.Chat, t string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: c.ID,
		Text:   t,
	})
	if err != nil {
		h.log.Error("[bot] failed to send message", "error", err.Error(), "chat", logChat(&c))
	}
	h.log.Info("[bot] message sent", "chat", logChat(&c))
}
