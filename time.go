package time

import (
	"context"
)

// This file provides implementations of functions provided by the time package but which take a
// context argument.  This makes it easier to identify when a function is being used in a
// context-aware way (or is incorrectly using non-context aware functions), and also allows for
// mocking, cancellation and timeouts, where appropriate.
//
// To avoid repeatedly inspecting the context, when making repeated use of these functions it is
// recommended to explicitly obtain the context Clock to be used directly:
//
//   var c time.Clock = time.ClockFromContext(ctx)
//
//   c.Now()
//   c.After(5 * time.Second)
//   ...

func AfterFunc(ctx context.Context, d Duration, f func()) *Timer {
	return ClockFromContext(ctx).AfterFunc(d, f)
}

func NewTicker(ctx context.Context, d Duration) *Ticker {
	return ClockFromContext(ctx).NewTicker(d)
}

func NewTimer(ctx context.Context, d Duration) *Timer {
	return ClockFromContext(ctx).NewTimer(d)
}

// Now returns the current time from the Clock in the given context. If there is no clock in the
// context, the real-time clock will be used.
func Now(ctx context.Context) Time {
	return ClockFromContext(ctx).Now()
}

func Tick(ctx context.Context, d Duration) <-chan Time {
	return ClockFromContext(ctx).Tick(d)
}

// Sleep suspends the calling goroutine for the duration specified.
//
// If the clock in the context is a mock clock, the duration of the sleep may be modified by the
// configuration of the mock.
func Sleep(ctx context.Context, d Duration) {
	ClockFromContext(ctx).Sleep(d)
}
