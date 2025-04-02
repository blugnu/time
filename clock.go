package time

import (
	"context"
	"time"
)

// Clock represents an interface described by the functions in the time package
// of the standard library.  It extends the time package with additional
// methods to create contexts with deadlines and timeouts based on the clock
// providing the interface.
//
// This allows for the creation of mock clocks for testing purposes through an
// API that is similar to and consistent with that of the system clock in the
// standard library `time` package.
type Clock interface {
	// After returns a channel that will send the current time after at least
	// duration d.
	After(d time.Duration) <-chan time.Time

	// AfterFunc waits for the duration to elapse and then calls f in its own
	// goroutine. It returns a Timer that can be used to stop the countdown
	// or to reset the Timer to run at a different time.
	//
	// If the Timer is stopped, the function f will not be called.
	AfterFunc(d time.Duration, f func()) *Timer

	// NewTicker returns a new Ticker that will send the current time on its
	// channel after each tick. The duration d must be greater than zero; if
	// d <= 0, NewTicker will panic.
	//
	// The duration of the Ticker can be modified using the Reset method
	//
	// The Ticker will continue ticking until Stop is called on it.
	NewTicker(d time.Duration) *Ticker

	// NewTimer returns a new Timer that will send the current time on its
	// channel after the duration d. The duration d must be greater than zero;
	// if d <= 0, NewTimer will panic.
	//
	// The duration of the Timer can be modified using the Reset method.
	//
	// The Timer will tick at the designated time unless Stop is called on it
	// beforehand.  A stopped Timer will resume if it is reset.
	NewTimer(d time.Duration) *Timer

	// Now returns the current time.
	Now() time.Time

	// Since returns the duration since t, according to the current time.  It is
	// shorthand for time.Since(c.Now()).
	Since(t time.Time) time.Duration

	// Sleep pauses the calling goroutine for at least the duration d.
	Sleep(d time.Duration)

	// Tick returns a channel that will send the current time after each tick.
	// Unlike NewTicker, if the duration d is zero or negative, Tick will return
	// a nil channel and will not panic.
	Tick(d time.Duration) <-chan time.Time

	// Until returns the duration until t, according to the current time.  It is
	// shorthand for time.Until(c.Now()).
	Until(t time.Time) time.Duration

	// ContextWithDeadline returns a new context with the given deadline. If the
	// given time is in the past, the returned context is already done.
	//
	// This function should be used in preference over context.WithDeadline
	// to ensure that code relying on the deadline behaves correctly under test
	// conditions which may provide a mock clock in the parent context.
	ContextWithDeadline(ctx context.Context, d time.Time) (context.Context, context.CancelFunc)

	// ContextWithDeadlineCause returns a new context with the given deadline and
	// cause. If the given time is in the past, the returned context is already
	// done.
	//
	// The cause is used to set the context error.
	//
	// This function should be used in preference over context.WithDeadlineCause
	// to ensure that code relying on the deadline behaves correctly under test
	// conditions which may provide a mock clock in the parent context.
	ContextWithDeadlineCause(ctx context.Context, d time.Time, cause error) (context.Context, context.CancelFunc)

	// ContextWithTimeout returns a new context with the given timeout. If the
	// given duration is zero or negative, the returned context is already done.
	//
	// This function should be used in preference over context.WithTimeout
	// to ensure that code relying on the timeout behaves correctly under test
	// conditions which may provide a mock clock in the parent context.
	ContextWithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc)

	// ContextWithTimeoutCause returns a new context with the given timeout and
	// cause. If the given duration is zero or negative, the returned context is
	// already done.
	//
	// The cause is used to set the context error.
	//
	// This function should be used in preference over context.WithTimeoutCause
	// to ensure that code relying on the timeout behaves correctly under test
	// conditions which may provide a mock clock in the parent context.
	ContextWithTimeoutCause(ctx context.Context, d time.Duration, cause error) (context.Context, context.CancelFunc)
}

// sysClock is a Clock implementation that wraps the system clock.
var sysClock Clock = systemClock{}

// SystemClock returns a clock implementation that uses the `time` package functions of the
// standard library.
func SystemClock() Clock {
	return sysClock
}

type systemClock struct{}

func (c systemClock) After(d time.Duration) <-chan Time { return time.After(d) }
func (c systemClock) AfterFunc(d time.Duration, f func()) *Timer {
	return &Timer{Timer: time.AfterFunc(d, f)}
}
func (c systemClock) Now() time.Time                        { return time.Now() }
func (c systemClock) Since(t time.Time) time.Duration       { return time.Since(t) }
func (c systemClock) Until(t time.Time) time.Duration       { return time.Until(t) }
func (c systemClock) Sleep(d time.Duration)                 { time.Sleep(d) }
func (c systemClock) Tick(d time.Duration) <-chan time.Time { return time.Tick(d) }

func (c systemClock) NewTicker(d time.Duration) *Ticker {
	return &Ticker{Ticker: time.NewTicker(d), initialised: true}
}

func (c systemClock) NewTimer(d time.Duration) *Timer {
	return &Timer{Timer: time.NewTimer(d), initialised: true}
}

func (c systemClock) ContextWithDeadline(ctx context.Context, d time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, d)
}

func (c systemClock) ContextWithDeadlineCause(ctx context.Context, d time.Time, cause error) (context.Context, context.CancelFunc) {
	return context.WithDeadlineCause(ctx, d, cause)
}

func (c systemClock) ContextWithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

func (c systemClock) ContextWithTimeoutCause(ctx context.Context, d time.Duration, cause error) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(ctx, d, cause)
}
