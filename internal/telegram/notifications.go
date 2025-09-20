package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/locales"
)

// Notifications handles sending notifications to users about process updates.
type Notifications struct {
	log *slog.Logger
	bot *bot.Bot
	in  <-chan entities.Process
}

// NewNotifications creates a new Notifications instance.
func NewNotifications(log *slog.Logger, bot *bot.Bot, in <-chan entities.Process) *Notifications {
	return &Notifications{
		log: log,
		bot: bot,
		in:  in,
	}
}

// Start begins listening for process updates and sending notifications.
func (n *Notifications) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				n.log.Info("[bot] notifications stopped")
				return
			case process := <-n.in:
				msg := n.getMessage(process)
				n.replyMessage(ctx, process, msg)
			}
		}
	}()
}

func (n *Notifications) getMessage(process entities.Process) string {
	switch {
	case process.Step == entities.StepDownloading && process.Status == entities.StatusInProgress:
		return locales.MsgDownloadStarted
	case process.Status == entities.StatusSuccess:
		return n.getSuccessMessage(process)
	case process.Status == entities.StatusFailed:
		return msgErr(process.Error)
	default:
		return ""
	}
}

func (n *Notifications) getSuccessMessage(process entities.Process) string {
	var title string
	if process.Episode != nil {
		title = process.Episode.Title
	}
	return fmt.Sprintf(locales.MsgDownloadSuccess, title)
}

func (n *Notifications) replyMessage(ctx context.Context, process entities.Process, text string) {
	if text == "" {
		// Ignore
		return
	}
	params := &bot.SendMessageParams{
		ChatID:          process.Request.ChatID,
		Text:            text,
		ReplyParameters: &models.ReplyParameters{MessageID: process.Request.MessageID},
	}

	msg, err := n.bot.SendMessage(ctx, params)
	if err != nil {
		n.log.Error("[bot] failed to send notification text",
			"error", err, "request", process.Request.LogValue(), "message", logMessage(msg))
	}
	n.log.Info("[bot] notification text sent",
		"request", process.Request.LogValue(), "message", logMessage(msg))
}
