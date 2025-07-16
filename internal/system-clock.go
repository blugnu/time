package internal

import (
	"context"
	"time"
)

var SystemClockInstance Clock = SystemClock{}

// SystemClock implements the Clock interface using the standard time package.
// It provides methods to interact with the system clock, such as getting the current time,
// sleeping for a duration, and creating timers and tickers.
type SystemClock struct{}

func (c SystemClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (c SystemClock) AfterFunc(d time.Duration, f func()) *Timer {
	return &Timer{Timer: time.AfterFunc(d, f), initialised: true}
}
func (c SystemClock) Now() time.Time                        { return time.Now() }
func (c SystemClock) Since(t time.Time) time.Duration       { return time.Since(t) }
func (c SystemClock) Until(t time.Time) time.Duration       { return time.Until(t) }
func (c SystemClock) Sleep(d time.Duration)                 { time.Sleep(d) }
func (c SystemClock) Tick(d time.Duration) <-chan time.Time { return time.Tick(d) }

func (c SystemClock) NewTicker(d time.Duration) *Ticker {
	return &Ticker{Ticker: time.NewTicker(d), initialised: true}
}

func (c SystemClock) NewTimer(d time.Duration) *Timer {
	return &Timer{Timer: time.NewTimer(d), initialised: true}
}

func (c SystemClock) ContextWithDeadline(ctx context.Context, d time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, d)
}

func (c SystemClock) ContextWithDeadlineCause(ctx context.Context, d time.Time, cause error) (context.Context, context.CancelFunc) {
	return context.WithDeadlineCause(ctx, d, cause)
}

func (c SystemClock) ContextWithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

func (c SystemClock) ContextWithTimeoutCause(ctx context.Context, d time.Duration, cause error) (context.Context, context.CancelFunc) {
	return context.WithTimeoutCause(ctx, d, cause)
}
