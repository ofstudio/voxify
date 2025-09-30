package telegram

import (
	"context"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Middleware defines a function to process middleware.
type Middleware func(next HandlerFunc) HandlerFunc

// HandlerFunc defines a function to handle a Telegram update.
type HandlerFunc func(ctx context.Context, bot Bot, update *models.Update)

// Bot is a minimal abstraction over github.com/go-telegram/bot.
// It defines the subset of functionality needed in this project and enables
// replacing the real bot with a mock for convenient testing.
type Bot interface {
	// RegisterHandler registers a new handler with a string pattern.
	RegisterHandler(handlerType bot.HandlerType, pattern string, matchType bot.MatchType, f HandlerFunc, m ...Middleware) string

	// RegisterHandlerRegexp registers a new handler with a regexp pattern.
	RegisterHandlerRegexp(handlerType bot.HandlerType, re *regexp.Regexp, f HandlerFunc, m ...Middleware) string

	// RegisterHandlerMatchFunc registers a new handler with a custom match function.
	RegisterHandlerMatchFunc(matchFunc bot.MatchFunc, f HandlerFunc, m ...Middleware) string

	// UnregisterHandler removes a previously registered handler by its ID.
	UnregisterHandler(id string)

	// SendMessage sends a message to a chat and returns the created message.
	SendMessage(ctx context.Context, p *bot.SendMessageParams) (*models.Message, error)

	// ProcessUpdate passes a single update through the bot’s handler chain.
	ProcessUpdate(ctx context.Context, upd *models.Update)

	// Start begins receiving updates and dispatching them to handlers.
	Start(ctx context.Context)

	// ID returns the bot's Telegram ID.
	ID() int64
}
