// Package telegram provides a tiny, testable wrapper around github.com/go-telegram/bot.
//
// It defines a minimal Bot interface (register handlers, send messages, process
// single updates, start polling, get ID) so application code depends on an
// interface and can be unit‑tested with mocks. NewBot returns an adapter that
// forwards calls to the real *bot.Bot while adapting handler + middleware
// signatures.
package telegram
