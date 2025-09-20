package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Context returns a context that is canceled when the process receives
// one of the specified signals: SIGINT, SIGTERM, SIGQUIT.
//
// Returns a context and a cancel function.
//
// If a callback is provided, it will be called with the received signal
// before the context is canceled.
func Context(ctx context.Context, callback ...func(os.Signal)) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		s := <-stop
		if len(callback) > 0 {
			callback[0](s)
		}
		cancel()
	}()
	return ctx, cancel
}
