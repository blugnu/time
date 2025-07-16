package internal

import "context"

type contextKey int

const clockKey contextKey = iota

// ContextWithClock returns a new context containing a specified clock
func ContextWithClock(ctx context.Context, c Clock) context.Context {
	return context.WithValue(ctx, clockKey, c)
}

// ClockFromContext returns the Clock in the specified context.  Returns
// nil if no Clock is present.
func ClockFromContext(ctx context.Context) Clock {
	if clock, ok := ctx.Value(clockKey).(Clock); ok {
		return clock
	}
	return nil
}
