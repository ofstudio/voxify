// Package cancelgroup provides a WaitGroup-like synchronization primitive
// that also supports cancellation via context.Context.
//
// It allows you to run goroutines that can be canceled as a group,
// making it easier to coordinate shutdown and resource cleanup.
package cancelgroup

import (
	"context"
	"sync"
)

// Group is similar to sync.WaitGroup but comes with a cancellable context.
// It can be used to launch goroutines, wait for their completion, and
// cancel them all at once.
type Group struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Group with its own cancellable context.
func New() *Group {
	ctx, cancel := context.WithCancel(context.Background())
	return &Group{
		ctx:    ctx,
		cancel: cancel,
	}
}

// WithContext creates a new Group with a parent context.
// Cancelling the parent will also cancel the group.
func WithContext(parent context.Context) *Group {
	ctx, cancel := context.WithCancel(parent)
	return &Group{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Add adds delta to the Group's counter (same as sync.WaitGroup.Add).
func (g *Group) Add(delta int) {
	g.wg.Add(delta)
}

// Done decrements the Group's counter (same as sync.WaitGroup.Done).
func (g *Group) Done() {
	g.wg.Done()
}

// Wait blocks until the Group's counter goes back to zero.
func (g *Group) Wait() {
	g.wg.Wait()
}

// Cancel signals cancellation to all goroutines launched under this group.
func (g *Group) Cancel() {
	g.cancel()
}

// Context returns the cancellable context associated with the group.
func (g *Group) Context() context.Context {
	return g.ctx
}

// Go runs the given function in a new goroutine and tracks it in the group.
// It's a convenience method equivalent to calling Add(1) before launching
// the goroutine and Done() when it finishes.
//
// The function f should respect the group's context for cancellation
// and must not panic.
func (g *Group) Go(f func()) {
	g.Add(1)
	go func() {
		defer g.Done()
		f()
	}()
}
