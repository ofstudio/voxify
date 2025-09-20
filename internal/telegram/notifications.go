package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/locales"
	"github.com/ofstudio/voxify/internal/services"
)

// Notifications handles sending notifications to users about process updates.
type Notifications struct {
	bot      *bot.Bot
	log      *slog.Logger
	notifier Notifier
}

// NewNotifications creates a new Notifications instance.
func NewNotifications(bot *bot.Bot, notifier Notifier, log *slog.Logger) *Notifications {
	return &Notifications{
		bot:      bot,
		log:      log,
		notifier: notifier,
	}
}

// Start begins listening for process updates and sending notifications.
func (n *Notifications) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				n.log.Info("bot notifications stopped")
				return
			case process := <-n.notifier.Notify():
				msg := n.getMessage(process)
				n.sendMessage(ctx, process, msg)
			}
		}
	}()
}

func (n *Notifications) getMessage(process *entities.Process) string {
	switch {
	case process.Step == entities.StepDownloading && process.Status == entities.StatusInProgress:
		return n.getDownloadStartedMessage()
	case process.Status == entities.StatusSuccess:
		return n.getSuccessMessage(process)
	case process.Status == entities.StatusFailed:
		return n.getErrorMessage(process)
	default:
		return ""
	}
}

func (n *Notifications) getDownloadStartedMessage() string {
	return locales.MsgDownloadStarted
}

func (n *Notifications) getSuccessMessage(process *entities.Process) string {
	var title string
	if process.Episode != nil {
		title = process.Episode.Title
	}
	return fmt.Sprintf(locales.MsgDownloadSuccess, title)
}

func (n *Notifications) getErrorMessage(process *entities.Process) string {
	if process.Error == nil {
		return locales.MsgSomethingWentWrong
	}

	// Init if error is services.Error
	var servicesErr services.Error
	if errors.As(process.Error, &servicesErr) {
		return n.getErrorMessageByCode(servicesErr.Code)
	}

	// Generic error message for non-services errors
	return locales.MsgSomethingWentWrong
}

func (n *Notifications) getErrorMessageByCode(code int) string {
	// Handle specific error codes in range 100-199
	switch code {
	case 101: // ErrNoMatchingPlatform
		return locales.MsgNoMatchingPlatform
	case 102: // ErrDownloadFailed
		return locales.MsgDownloadFailed
	case 103: // ErrEpisodeInProgress
		return locales.MsgEpisodeInProgress
	case 104: // ErrEpisodeExists
		return locales.MsgEpisodeExists
	case 105: // ErrProcessInterrupted
		return locales.MsgProcessInterrupted
	default:
		return fmt.Sprintf(locales.MsgSomethingWentWrongWithCode, code)
	}
}

func (n *Notifications) sendMessage(ctx context.Context, process *entities.Process, text string) {
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
