package events

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/ofstudio/voxify/internal/domain"
)

// AsyncBus is an asynchronous in-memory domain.EventBus implementation.
// It dispatches events to multiple subscribed handlers concurrently in goroutines.
// Safe for concurrent use.
type AsyncBus struct {
	log         *slog.Logger
	handlersMap map[domain.EventType][]domain.EventHandler // mapping event type → handlers
	mu          sync.RWMutex                               // protects handlersMap
	wg          sync.WaitGroup                             // tracks active handler goroutines
}

// NewAsyncBus creates a new AsyncBus instance.
func NewAsyncBus(log *slog.Logger) *AsyncBus {
	return &AsyncBus{
		log:         log,
		handlersMap: make(map[domain.EventType][]domain.EventHandler),
	}
}

// Publish dispatches the given event to all subscribed handlers asynchronously.
// Each handler is executed in its own goroutine.
func (b *AsyncBus) Publish(event domain.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if handlers, ok := b.handlersMap[event.Type()]; ok {
		for _, handler := range handlers {
			b.wg.Add(1) // Move this before goroutine launch to avoid race condition
			go func(handler domain.EventHandler, event domain.Event) {
				defer b.wg.Done()
				defer b.handlerRecover(event)
				handler(event)
			}(handler, event)
		}
	}
}

// handlerRecover recovers from panics in handlers and logs the error.
func (b *AsyncBus) handlerRecover(event domain.Event) {
	if r := recover(); r != nil {
		b.log.Error("[event bus] handler error",
			"error", fmt.Errorf("handler panic: %v", r),
			"event", event,
		)
	}
}

// Subscribe registers a new handler for the given event type.
// Multiple handlers can be subscribed to the same type.
func (b *AsyncBus) Subscribe(t domain.EventType, h domain.EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlersMap[t] = append(b.handlersMap[t], h)
}

// Wait blocks until all currently running event handlers have finished.
// Useful for graceful shutdown.
func (b *AsyncBus) Wait() {
	b.wg.Wait()
}
