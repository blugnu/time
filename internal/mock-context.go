package internal

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ensure that mockContext implements the context.Context interface
var _ context.Context = (*MockContext)(nil)

// MockContext is a mock implementation of context.Context that uses a
// MockClock to simulate the passage of time. It allows for testing of
// context cancellation and deadlines without relying on real time.
type MockContext struct {
	sync.Mutex

	clock    Clock
	parent   context.Context //nolint:containedctx // needed to implement a mock context
	deadline time.Time
	cause    error
	done     chan struct{}

	err   error
	timer *Timer
}

// NewMockContext returns a new context with the given deadline and a
// cancellable timer.
//
// If the specified deadline has already passed, the context is immediately
// cancelled with context.DeadlineExceeded.
func NewMockContext(
	parent context.Context,
	clock Clock,
	deadline time.Time,
	cause error,
) (*MockContext, context.CancelFunc) {
	ctx := &MockContext{
		clock:    clock,
		parent:   parent,
		deadline: deadline.UTC(),
		done:     make(chan struct{}),
		cause:    cause,
	}

	// if the parent has a cancellation channel arrange to cancel the new
	// child context if the parent is cancelled
	if parent.Done() != nil {
		go func() {
			select {
			case <-parent.Done():
				ctx.cancel(parent.Err())
			case <-ctx.Done():
				// if the child context is cancelled, stop listening for
				// cancellation on the parent context
			}
		}()
	}

	dur := clock.Until(deadline)
	if dur <= 0 {
		ctx.cancel(context.DeadlineExceeded) // deadline has already passed
		return ctx, func() { /* NO-OP */ }
	}

	ctx.Lock()
	defer ctx.Unlock()

	if ctx.err == nil {
		// if the context is not already cancelled, start a timer to cancel
		// the context when the deadline is reached, wrapping the cause
		// error in context.DeadlineExceeded if it is not nil
		ctx.timer = clock.AfterFunc(dur, func() {
			err := context.DeadlineExceeded
			if cause != nil {
				err = fmt.Errorf("%w: %w", err, cause)
			}
			ctx.cancel(err)
		})
	}

	// return the new context and a cancel function
	// the cancel function will stop the timer if it is still running
	// and cancel the context
	return ctx, func() { ctx.cancel(context.Canceled) }
}

func (c *MockContext) String() string {
	remaining := c.clock.Until(c.deadline)
	s := fmt.Sprintf("MockContext{deadline: %s (in %s)", c.deadline, remaining)

	switch {
	case c.err != nil:
		s += fmt.Sprintf(", err: %q", c.err)
	case c.cause != nil:
		s += fmt.Sprintf(", cause: %q", c.cause)
	}

	return s + "}"
}

func (c *MockContext) Deadline() (time.Time, bool) { return c.deadline, true }

func (c *MockContext) Done() <-chan struct{} { return c.done }

func (c *MockContext) Err() error { return c.err }

func (c *MockContext) Value(key any) any { return c.parent.Value(key) }

func (c *MockContext) cancel(err error) {
	c.Lock()
	defer c.Unlock()

	if c.err != nil {
		return // already canceled
	}

	c.err = err
	close(c.done)

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
}
