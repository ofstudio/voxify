package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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

// CmdStart handles the /start command.
func (h *Handlers) CmdStart() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		h.log.Info("[bot] start command received", "update_id", update.ID, "message", logMessage(update.Message))
		h.sendMessage(ctx, b, update.Message.Chat, locales.MsgStart)
	}
}

// CmdBuild handles the /build command to manually trigger RSS feed rebuild.
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

// CmdInfo handles the /info command to provide information about the podcast feed
func (h *Handlers) CmdInfo() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		h.log.Info("[bot] info command received", "update_id", update.ID, "message", logMessage(update.Message))

		msg, err := h.getInfoMessage(ctx)
		if err != nil {
			msg = msgErr(err)
		}

		h.sendMessage(ctx, b, update.Message.Chat, msg, models.ParseModeHTML)
	}
}

// getInfoMessage retrieves podcast feed information and formats it into a message.
// Example output:
//
//	🏷️ Title:
//	My podcast title
//
//	💬 This is description
//
//	👨‍💻 By Oleg Fomin
//	🌐 Language: en
//	📚 Categories: Science, Physics
//	🔑 Keywords: Tech,Talks,News
//	🖼️ Artwork <- link to image
//	🔗 Website <- link to website
//	🎧 Number of episodes: 12
//	🔞 Explicit content
//
//	📡 RSS: https://example.com/feed.rss
func (h *Handlers) getInfoMessage(ctx context.Context) (string, error) {
	// Retrieve feed info
	feed, err := h.feeder.Feed(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get feed info: %w", err)
	}

	// Basic info: title, description
	msg := fmt.Sprintf(locales.MsgFeedInfoBasic, feed.Title, feed.Description)
	// Author
	if feed.Author != "" {
		msg += fmt.Sprintf(locales.MsgFeedInfoAuthor, feed.Author)
	}
	// Language, categories
	msg += fmt.Sprintf(locales.MsgFeedInfoLanguage, feed.Language)
	// Categories
	msg += fmt.Sprintf(locales.MsgFeedInfoCategories, categoriesToString(feed.Categories))
	// Keywords
	if feed.Keywords != "" {
		msg += fmt.Sprintf(locales.MsgFeedInfoKeywords, feed.Keywords)
	}
	// Artwork
	if feed.ImageUrl != "" {
		msg += fmt.Sprintf(locales.MsgFeedInfoArtwork, feed.ImageUrl)
	}
	// Website
	if feed.WebsiteLink != "" {
		msg += fmt.Sprintf(locales.MsgFeedInfoWebsite, feed.WebsiteLink)
	}
	// Episodes
	if feed.EpisodeCount > 0 {
		msg += fmt.Sprintf(locales.MsgFeedInfoEpisodes, feed.EpisodeCount)
	} else {
		msg += locales.MsgFeedInfoNoEpisodes
	}
	// Explicit
	if feed.Explicit {
		msg += locales.MsgFeedInfoExplicit
	}
	// RSS link
	msg += fmt.Sprintf(locales.MsgFeedInfoRSS, feed.RSSLink)

	return msg, nil
}

// categoriesToString converts a slice of FeedCategory to a comma-separated string.
// It includes subcategories as well.
func categoriesToString(categories []entities.FeedCategory) string {
	var cats []string
	for _, c := range categories {
		cats = append(cats, c.Text)
		if len(c.Subcategories) > 0 {
			cats = append(cats, c.Subcategories...)
		}
	}
	return strings.Join(cats, ", ")
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
func (h *Handlers) sendMessage(ctx context.Context, b *bot.Bot, c models.Chat, t string, m ...models.ParseMode) {
	var mode models.ParseMode
	if len(m) > 0 {
		mode = m[0]
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    c.ID,
		Text:      t,
		ParseMode: mode,
	})
	if err != nil {
		h.log.Error("[bot] failed to send message", "error", err.Error(), "chat", logChat(&c))
	}
	h.log.Info("[bot] message sent", "chat", logChat(&c))
}
