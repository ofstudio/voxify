package handlers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/templates"
	"github.com/ofstudio/voxify/pkg/telegram"
)

// NotificationHandlers handles sending Telegram notifications for response events.
type NotificationHandlers struct {
	log *slog.Logger
	ctx context.Context
	bus domain.EventBus
	api telegram.API
}

func NewNotificationHandlers(log *slog.Logger, bus domain.EventBus, api telegram.API) *NotificationHandlers {
	return &NotificationHandlers{log: log, bus: bus, api: api}
}

// Start initializes the notification handlers.
func (h *NotificationHandlers) Start(cxt context.Context) {
	// subscribe to events
	h.ctx = cxt
	h.bus.Subscribe(domain.DownloadResponseEvent, h.downloadNotificationHandler)
	h.bus.Subscribe(domain.BuildResponseEvent, h.buildNotificationHandler)
	h.bus.Subscribe(domain.FeedInfoResponseEvent, h.feedInfoNotificationHandler)
}

// downloadNotificationHandler handles download response events and sends notifications.
func (h *NotificationHandlers) downloadNotificationHandler(event domain.Event) {
	resp := event.Payload().(domain.DownloadResponse)
	switch resp.Status {
	case domain.StatusPending:
		h.sendReply(templates.MsgDownloadStarted, resp.Request.Source)
	case domain.StatusSuccess:
		h.sendReply(templates.MsgDownloadSuccess(resp.Episode.Title), resp.Request.Source)
	case domain.StatusFailed:
		h.sendReply(templates.MsgError(resp.Error), resp.Request.Source)
	default:
		h.log.Error("[notification handlers] unknown download response status",
			"response", resp)
		h.sendReply(templates.MsgSomethingWentWrong, resp.Request.Source)
	}
}

// buildNotificationHandler handles build response events and sends notifications.
func (h *NotificationHandlers) buildNotificationHandler(event domain.Event) {
	resp := event.Payload().(domain.BuildResponse)
	// Skip if non-user initiated build
	if resp.Request.Source == nil {
		return
	}
	switch resp.Status {
	case domain.StatusPending:
		// no-op
	case domain.StatusSuccess:
		h.sendReply(templates.MsgBuildSuccess, *resp.Request.Source)
	case domain.StatusFailed:
		h.sendReply(templates.MsgError(resp.Error), *resp.Request.Source)
	default:
		h.log.Error("[notification handlers] unknown build response status",
			"response", resp)
		h.sendReply(templates.MsgSomethingWentWrong, *resp.Request.Source)
	}
}

// feedInfoNotificationHandler handles feed info response events and sends notifications.
func (h *NotificationHandlers) feedInfoNotificationHandler(event domain.Event) {
	resp := event.Payload().(domain.FeedInfoResponse)
	switch resp.Status {
	case domain.StatusPending:
		// no-op
	case domain.StatusSuccess:
		b := &strings.Builder{}
		if err := templates.FeedInfoTemplate.Execute(b, resp.FeedInfo); err != nil {
			h.log.Error("[telegram handlers] failed to execute feed info template", "error", err.Error())
			h.sendReply(templates.MsgSomethingWentWrong, resp.Request.Source)
			return
		}
		h.sendReply(b.String(), resp.Request.Source)
	case domain.StatusFailed:
		h.sendReply(templates.MsgError(resp.Error), resp.Request.Source)
	default:
		h.log.Error("[notification handlers] unknown feed info response status",
			"response", resp)
		h.sendReply(templates.MsgSomethingWentWrong, resp.Request.Source)
	}
}

// sendReply sends a reply message to the user.
func (h *NotificationHandlers) sendReply(msg string, src domain.RequestSource) {
	h.sendMessage(&bot.SendMessageParams{
		ChatID:          src.ChatID,
		Text:            msg,
		ParseMode:       "HTML",
		ReplyParameters: &models.ReplyParameters{MessageID: src.MessageID},
	})
}

// sendMessage sends a message via the Telegram bot and logs the result.
func (h *NotificationHandlers) sendMessage(params *bot.SendMessageParams) {
	msg, err := h.api.SendMessage(h.ctx, params)
	if err != nil {
		h.log.Error("[telegram handlers] failed to send message",
			"error", err.Error(), "chat_id", params.ChatID)
		return
	}
	h.log.Info("[notification handlers] message sent",
		"message", telegram.LogMessage(msg),
	)
}
