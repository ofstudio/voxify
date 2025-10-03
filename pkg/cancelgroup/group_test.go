package cancelgroup

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestGroupWait(t *testing.T) {
	g := New()

	var counter int32

	g.Go(func() {
		atomic.AddInt32(&counter, 1)
	})

	g.Wait()

	if counter != 1 {
		t.Errorf("expected counter=1, got %d", counter)
	}
}

func TestGroupCancel(t *testing.T) {
	g := New()

	stopped := make(chan struct{})

	g.Go(func() {
		select {
		case <-g.Context().Done():
			close(stopped)
		}
	})

	// Cancel should notify goroutines
	g.Cancel()

	select {
	case <-stopped:
		// ok
	case <-time.After(5 * time.Second):
		t.Fatal("expected goroutine to stop after cancel")
	}

	g.Wait()
}

func TestWithContext(t *testing.T) {
	parent, cancel := New().Context(), func() {}

	// manually wrap parent into a context we can cancel
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	g := WithContext(ctx)

	stopped := make(chan struct{})
	g.Go(func() {
		<-g.Context().Done()
		close(stopped)
	})

	// cancel parent
	cancel()

	select {
	case <-stopped:
		// ok
	case <-time.After(5 * time.Second):
		t.Fatal("expected goroutine to stop after parent cancel")
	}

	g.Wait()
}
