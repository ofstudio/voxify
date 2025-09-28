package events

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/domain"
)

func TestAsyncBus(t *testing.T) {
	suite.Run(t, new(TestAsyncBusSuite))
}

type TestAsyncBusSuite struct {
	suite.Suite
	bus *AsyncBus
}

func (suite *TestAsyncBusSuite) SetupSubTest() {
	// Create a new AsyncBus instance for each subtest to ensure isolation
	suite.bus = NewAsyncBus(slog.Default())
}

func (suite *TestAsyncBusSuite) TestSubscribe() {
	suite.Run("Single handler subscription", func() {
		called := false
		handler := func(event domain.Event) {
			called = true
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, handler)

		// Verify handler was added by publishing an event and checking if it was called
		suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
			ID: "test",
		}))
		suite.bus.Wait()

		suite.True(called)
	})

	suite.Run("Multiple handlers for same event type", func() {
		var callCount atomic.Int32

		handler1 := func(event domain.Event) {
			callCount.Add(1)
		}
		handler2 := func(event domain.Event) {
			callCount.Add(1)
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, handler1)
		suite.bus.Subscribe(domain.DownloadRequestEvent, handler2)

		suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
			ID: "test",
		}))
		suite.bus.Wait()

		suite.Equal(int32(2), callCount.Load())
	})

	suite.Run("Handlers for different event types", func() {
		var downloadCalled, buildCalled bool

		downloadHandler := func(event domain.Event) {
			downloadCalled = true
		}
		buildHandler := func(event domain.Event) {
			buildCalled = true
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, downloadHandler)
		suite.bus.Subscribe(domain.BuildRequestEvent, buildHandler)

		// Publish download event
		suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
			ID: "test",
		}))
		suite.bus.Wait()

		suite.True(downloadCalled)
		suite.False(buildCalled)

		// Reset and publish build event
		downloadCalled = false
		suite.bus.Publish(domain.NewBuildRequestEvent(domain.BuildRequest{
			ID: "test",
		}))
		suite.bus.Wait()

		suite.False(downloadCalled)
		suite.True(buildCalled)
	})
}

func (suite *TestAsyncBusSuite) TestPublish() {
	suite.Run("Publish to subscribed handlers", func() {
		var receivedEvent domain.Event
		var wg sync.WaitGroup
		wg.Add(1)

		handler := func(event domain.Event) {
			defer wg.Done()
			receivedEvent = event
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, handler)

		testEvent := domain.NewDownloadRequestEvent(domain.DownloadRequest{
			ID:  "test-123",
			Url: "https://example.com",
		})

		suite.bus.Publish(testEvent)
		wg.Wait()

		suite.Equal(testEvent.Type(), receivedEvent.Type())
		suite.Equal(testEvent.Payload(), receivedEvent.Payload())
	})

	suite.Run("Publish to non-existent event type", func() {
		// Should not cause any issues
		suite.NotPanics(func() {
			suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
				ID: "test",
			}))
			suite.bus.Wait()
		})
	})

	suite.Run("Publish with no handlers", func() {
		// Should not cause any issues when no handlers are subscribed
		suite.NotPanics(func() {
			suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
				ID: "test",
			}))
			suite.bus.Wait()
		})
	})
}

func (suite *TestAsyncBusSuite) TestWait() {
	suite.Run("Wait for all handlers to complete", func() {
		var counter atomic.Int32
		const numHandlers = 5

		for i := 0; i < numHandlers; i++ {
			handler := func(event domain.Event) {
				time.Sleep(50 * time.Millisecond) // Simulate work
				counter.Add(1)
			}
			suite.bus.Subscribe(domain.DownloadRequestEvent, handler)
		}

		// Publish event and wait
		start := time.Now()
		suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
			ID: "test",
		}))
		suite.bus.Wait()
		elapsed := time.Since(start)

		// All handlers should have completed
		suite.Equal(int32(numHandlers), counter.Load())
		// Should take at least 50ms but less than 250ms (if truly parallel)
		suite.GreaterOrEqual(elapsed, 50*time.Millisecond)
		suite.Less(elapsed, 250*time.Millisecond)
	})

	suite.Run("Wait with no active handlers", func() {
		// Should return immediately
		start := time.Now()
		suite.bus.Wait()
		elapsed := time.Since(start)

		suite.Less(elapsed, 10*time.Millisecond)
	})
}

func (suite *TestAsyncBusSuite) TestConcurrentOperations() {
	suite.Run("Concurrent subscribe and publish", func() {
		var counter atomic.Int32
		const numGoroutines = 10
		var wg sync.WaitGroup

		// Start goroutines that subscribe and publish concurrently
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				// Subscribe handler
				handler := func(event domain.Event) {
					counter.Add(1)
				}
				suite.bus.Subscribe(domain.DownloadRequestEvent, handler)

				// Publish event
				suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
					ID: string(rune(id)),
				}))
			}(i)
		}

		wg.Wait()
		suite.bus.Wait() // Wait for all event handlers to complete

		// Should not panic and counter should be greater than 0
		suite.Greater(counter.Load(), int32(0))
	})

	suite.Run("Multiple concurrent publishes", func() {
		var counter atomic.Int32

		handler := func(event domain.Event) {
			counter.Add(1)
		}
		suite.bus.Subscribe(domain.DownloadRequestEvent, handler)

		const numPublishes = 100
		var wg sync.WaitGroup

		for i := 0; i < numPublishes; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
					ID: "test",
				}))
			}()
		}

		wg.Wait()
		suite.bus.Wait()

		suite.Equal(int32(numPublishes), counter.Load())
	})
}

func (suite *TestAsyncBusSuite) TestPanicHandling() {
	suite.Run("Handler panic recovery", func() {
		var normalHandlerCalled bool

		// Handler that panics
		panicHandler := func(event domain.Event) {
			panic("test panic in handler")
		}

		// Normal handler that should still execute
		normalHandler := func(event domain.Event) {
			normalHandlerCalled = true
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, panicHandler)
		suite.bus.Subscribe(domain.DownloadRequestEvent, normalHandler)

		// Should not panic the test
		suite.NotPanics(func() {
			suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
				ID: "test",
			}))
			suite.bus.Wait()
		})

		// Normal handler should still be called despite panic in other handler
		suite.True(normalHandlerCalled)
	})

	suite.Run("Multiple panicking handlers", func() {
		const numPanicHandlers = 5

		for i := 0; i < numPanicHandlers; i++ {
			handler := func(event domain.Event) {
				panic("panic in handler")
			}
			suite.bus.Subscribe(domain.DownloadRequestEvent, handler)
		}

		// Should handle multiple panics gracefully
		suite.NotPanics(func() {
			suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
				ID: "test",
			}))
			suite.bus.Wait()
		})
	})

	suite.Run("Panic in slow handler", func() {
		var normalHandlerCalled bool

		// Handler that takes time then panics
		slowPanicHandler := func(event domain.Event) {
			time.Sleep(10 * time.Millisecond)
			panic("panic after delay")
		}

		// Normal handler that should still execute
		normalHandler := func(event domain.Event) {
			normalHandlerCalled = true
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, slowPanicHandler)
		suite.bus.Subscribe(domain.DownloadRequestEvent, normalHandler)

		suite.NotPanics(func() {
			suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
				ID: "test",
			}))
			suite.bus.Wait()
		})

		// Normal handler should still be called despite panic in other handler
		suite.True(normalHandlerCalled)
	})
}

func (suite *TestAsyncBusSuite) TestEdgeCases() {
	suite.Run("Subscribe with nil handler", func() {
		// Should not cause issues during subscription
		suite.NotPanics(func() {
			suite.bus.Subscribe(domain.DownloadRequestEvent, nil)
		})

		// Should not panic when trying to publish to nil handler (the implementation recovers from panics)
		suite.NotPanics(func() {
			suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
				ID: "test",
			}))
			suite.bus.Wait()
		})
	})

	suite.Run("Multiple Wait calls", func() {
		called := false
		handler := func(event domain.Event) {
			time.Sleep(20 * time.Millisecond)
			called = true
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, handler)
		suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
			ID: "test",
		}))

		// Multiple Wait calls should work correctly
		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				suite.bus.Wait()
			}()
		}

		wg.Wait()
		suite.True(called)
	})

	suite.Run("Handler runtime error", func() {
		var normalHandlerCalled bool

		// Handler that causes runtime error (nil pointer dereference)
		errorHandler := func(event domain.Event) {
			var p *int
			_ = *p // This will cause a runtime panic
		}

		// Normal handler that should still execute
		normalHandler := func(event domain.Event) {
			normalHandlerCalled = true
		}

		suite.bus.Subscribe(domain.DownloadRequestEvent, errorHandler)
		suite.bus.Subscribe(domain.DownloadRequestEvent, normalHandler)

		// Should not panic the test
		suite.NotPanics(func() {
			suite.bus.Publish(domain.NewDownloadRequestEvent(domain.DownloadRequest{
				ID: "test",
			}))
			suite.bus.Wait()
		})

		// Normal handler should still be called despite error in other handler
		suite.True(normalHandlerCalled)
	})
}
