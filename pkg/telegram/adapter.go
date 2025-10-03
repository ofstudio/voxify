package telegram

import (
	"context"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// adapter is a thin pass-through to *bot.Bot implementing the Bot interface.
type adapter struct{ b *bot.Bot }

// NewBot creates a new transparent adapter over *bot.Bot.
func NewBot(token string, opts ...bot.Option) (Bot, error) {
	b, err := bot.New(token, opts...)
	if err != nil {
		return nil, err
	}
	return &adapter{b: b}, nil
}

// RegisterHandler registers a new handler with a string pattern.
func (a *adapter) RegisterHandler(handlerType bot.HandlerType, pattern string, matchType bot.MatchType, f HandlerFunc, m ...Middleware) string {
	return a.b.RegisterHandler(handlerType, pattern, matchType, Handler(f), Middlewares(m...)...)
}

// RegisterHandlerRegexp registers a new handler with a regexp pattern.
func (a *adapter) RegisterHandlerRegexp(handlerType bot.HandlerType, re *regexp.Regexp, f HandlerFunc, m ...Middleware) string {
	return a.b.RegisterHandlerRegexp(handlerType, re, Handler(f), Middlewares(m...)...)
}

// RegisterHandlerMatchFunc registers a new handler with a custom match function.
func (a *adapter) RegisterHandlerMatchFunc(matchFunc bot.MatchFunc, f HandlerFunc, m ...Middleware) string {
	return a.b.RegisterHandlerMatchFunc(matchFunc, Handler(f), Middlewares(m...)...)
}

// UnregisterHandler removes a previously registered handler by its ID.
func (a *adapter) UnregisterHandler(id string) { a.b.UnregisterHandler(id) }

// SendMessage sends a message to a chat and returns the created message
func (a *adapter) SendMessage(ctx context.Context, p *bot.SendMessageParams) (*models.Message, error) {
	return a.b.SendMessage(ctx, p)
}

// ProcessUpdate passes a single update through the bot’s handler chain.
func (a *adapter) ProcessUpdate(ctx context.Context, upd *models.Update) {
	a.b.ProcessUpdate(ctx, upd)
}

// Start begins receiving updates and dispatching them to handlers.
func (a *adapter) Start(ctx context.Context) { a.b.Start(ctx) }

// ID returns the bot's Telegram ID.
func (a *adapter) ID() int64 { return a.b.ID() }

// Handler converts a HandlerFunc to a bot.HandlerFunc by injecting the Bot instance.
func Handler(f HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		f(ctx, &adapter{b: b}, update)
	}
}

// Middlewares converts a slice of Middleware to a slice of bot.Middleware by injecting the Bot instance.
func Middlewares(m ...Middleware) []bot.Middleware {
	bm := make([]bot.Middleware, 0, len(m))
	for _, mw := range m {
		bm = append(bm, func(next bot.HandlerFunc) bot.HandlerFunc {
			return Handler(mw(func(ctx context.Context, api API, update *models.Update) {
				next(ctx, api.(*adapter).b, update)
			}))
		})
	}
	return bm
}
