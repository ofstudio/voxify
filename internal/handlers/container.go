package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/pkg/telegram"
)

// Container holds all handler instances.
type Container struct {
	Request      *RequestHandlers
	Telegram     *TelegramHandlers
	Notification *NotificationHandlers
}

// New creates a new handler Container instance.
func New(cfg config.Settings, log *slog.Logger, bus domain.EventBus) *Container {
	return &Container{
		Request:      NewRequestHandlers(cfg, log, bus),
		Telegram:     NewTelegramHandlers(cfg, log, bus),
		Notification: NewNotificationHandlers(log, bus),
	}
}

// WithBot sets the bot instance.
func (c *Container) WithBot(bot telegram.Bot) *Container {
	c.Telegram.WithBot(bot)
	c.Notification.WithAPI(bot)
	return c
}

// WithBuilder sets the FeedBuilder instance.
func (c *Container) WithBuilder(b domain.FeedBuilder) *Container {
	c.Request.WithBuilder(b)
	return c
}

// WithDownloader sets the EpisodeDownloader instance.
func (c *Container) WithDownloader(d domain.EpisodeDownloader) *Container {
	c.Request.WithDownloader(d)
	return c
}

// Init initializes and starts all handlers.
func (c *Container) Init(ctx context.Context) error {
	if err := c.Request.Init(ctx); err != nil {
		return fmt.Errorf("failed to start event handlers: %w", err)
	}
	if err := c.Telegram.Init(ctx); err != nil {
		return fmt.Errorf("failed to start telegram handlers: %w", err)
	}

	if err := c.Notification.Init(ctx); err != nil {
		return fmt.Errorf("failed to start notification handlers: %w", err)
	}

	c.Telegram.log.Info("[handlers] handlers started")
	return nil
}

// Wait blocks until all handlers have completed their work.
func (c *Container) Wait() {
	c.Request.Wait()
}
