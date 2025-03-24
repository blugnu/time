package time

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type contextKey int

const clockKey contextKey = iota

// ClockFromContext returns the Clock in the given context.
// If no Clock is in the context the system clock is returned.
func ClockFromContext(ctx context.Context) Clock {
	if clock := TryClockFromContext(ctx); clock != nil {
		return clock
	}
	return SystemClock()
}

// TryClockFromContext returns the Clock in the given context or nil
// if no Clock is present.
func TryClockFromContext(ctx context.Context) Clock {
	if c, ok := ctx.Value(clockKey).(Clock); ok {
		return c
	}
	return nil
}

// ContextWithClock returns a new context containing a given clock.
//
//   - If the context already contains a clock the function panics with ErrClockAlreadyExists.
//   - If the given clock is nil a new context is returned with the system clock added.
func ContextWithClock(ctx context.Context, c Clock) context.Context {
	if clock := TryClockFromContext(ctx); clock != nil {
		panic(ErrClockAlreadyExists)
	}
	switch {
	case c == nil:
		return context.WithValue(ctx, clockKey, SystemClock())
	default:
		return context.WithValue(ctx, clockKey, c)
	}
}

// ContextWithDeadline returns a new context with the given deadline.
func ContextWithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return ContextWithDeadlineCause(parent, deadline, nil)
}

// ContextWithDeadlineCause returns a new context with the given deadline and cause.
func ContextWithDeadlineCause(parent context.Context, deadline time.Time, cause error) (context.Context, context.CancelFunc) {
	clock := ClockFromContext(parent)
	dur := clock.Until(deadline)
	return ContextWithTimeoutCause(parent, dur, cause)
}

// ContextWithTimeout returns a new context with the given timeout.
func ContextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return ContextWithTimeoutCause(parent, timeout, nil)
}

// ContextWithTimeoutCause returns a new context with the given timeout and cause.
//
// If the parent context has a deadline that occurs before the given timeout, a
// cancellable context is returned with the parent deadline. The cause is
// ignored in this case.
func ContextWithTimeoutCause(parent context.Context, timeout time.Duration, cause error) (context.Context, context.CancelFunc) {
	clock := ClockFromContext(parent)
	if _, isMock := clock.(*mockClock); !isMock {
		return context.WithTimeoutCause(parent, timeout, cause)
	}

	// if the parent context has a deadline which will occur before the timeout
	// return a cancellable context (inheriting the parent deadline since that will
	// be the effective timeout)
	deadline := clock.Now().Add(timeout)
	if parentDeadline, isMock := parent.Deadline(); isMock && parentDeadline.Before(deadline) {
		return context.WithCancel(parent)
	}

	return newMockContext(parent, clock, deadline)
}

// ContextWithMockClock returns a new context with a mock clock configured with the given options.
//
// If the parent context already has a clock the function panics with ErrClockAlreadyExists.
func ContextWithMockClock(parent context.Context, opts ...ClockOption) (context.Context, MockClock) {
	if clock := TryClockFromContext(parent); clock != nil {
		panic(ErrClockAlreadyExists)
	}

	m := NewMockClock(opts...)
	return ContextWithClock(parent, m), m
}

// ensure that mockContext implements the context.Context interface
var _ context.Context = (*mockContext)(nil)

type mockContext struct {
	sync.Mutex

	clock    Clock
	parent   context.Context
	deadline time.Time
	done     chan struct{}

	err   error
	timer *Timer
}

// newMockContext returns a new context with the given deadline and a
// cancellable timer.
//
// If the specified deadline has already passed, the context is immediately
// cancelled with context.DeadlineExceeded.
func newMockContext(
	parent context.Context,
	clock Clock,
	deadline time.Time,
) (*mockContext, context.CancelFunc) {
	ctx := &mockContext{
		clock:    clock,
		parent:   parent,
		deadline: deadline.UTC(),
		done:     make(chan struct{}),
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
		// the context when the deadline is reached
		ctx.timer = clock.AfterFunc(dur, func() {
			ctx.cancel(context.DeadlineExceeded)
		})
	}

	// return the new context and a cancel function
	// the cancel function will stop the timer if it is still running
	// and cancel the context
	return ctx, func() { ctx.cancel(context.Canceled) }
}

func (c *mockContext) cancel(err error) {
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

func (c *mockContext) Deadline() (deadline time.Time, ok bool) { return c.deadline, true }

func (c *mockContext) Done() <-chan struct{} { return c.done }

func (c *mockContext) Err() error { return c.err }

func (c *mockContext) Value(key any) any { return c.parent.Value(key) }

func (c *mockContext) String() string {
	return fmt.Sprintf("mock: context.WithDeadline: %s: %s", c.deadline.Sub(c.clock.Now()), c.deadline)
}
