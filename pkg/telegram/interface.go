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
type HandlerFunc func(ctx context.Context, api API, update *models.Update)

type Option func(b *Bot)

// Bot is a minimal abstraction over github.com/go-telegram/bot.
// It defines the subset of functionality needed in this project and enables
// replacing the real bot with a mock for convenient testing.
type Bot interface {
	API

	// ID returns the bot's Telegram ID.
	ID() int64

	// RegisterHandler registers a new handler with a string pattern.
	RegisterHandler(handlerType bot.HandlerType, pattern string, matchType bot.MatchType, f HandlerFunc, m ...Middleware) string

	// RegisterHandlerRegexp registers a new handler with a regexp pattern.
	RegisterHandlerRegexp(handlerType bot.HandlerType, re *regexp.Regexp, f HandlerFunc, m ...Middleware) string

	// RegisterHandlerMatchFunc registers a new handler with a custom match function.
	RegisterHandlerMatchFunc(matchFunc bot.MatchFunc, f HandlerFunc, m ...Middleware) string

	// UnregisterHandler removes a previously registered handler by its ID.
	UnregisterHandler(id string)

	// ProcessUpdate passes a single update through the bot’s handler chain.
	ProcessUpdate(ctx context.Context, upd *models.Update)

	// Start begins receiving updates and dispatching them to handlers.
	Start(ctx context.Context)
}

// API defines the subset of the Telegram Bot API  methods used in this project.
type API interface {
	// SendMessage sends a message to a chat and returns the created message.
	SendMessage(ctx context.Context, p *bot.SendMessageParams) (*models.Message, error)
}
