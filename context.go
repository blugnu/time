package time

import (
	"context"
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

// ContextWithMockClock returns a new context with a mock clock configured with the
// given options.  If the parent context already has a clock the function panics
// with ErrClockAlreadyExists.
//
// This function is provided as a convenience when writing tests requiring a mock
// clock.
func ContextWithMockClock(parent context.Context, opts ...ClockOption) (context.Context, MockClock) {
	if clock := TryClockFromContext(parent); clock != nil {
		panic(ErrClockAlreadyExists)
	}

	m := NewMockClock(opts...)
	return ContextWithClock(parent, m), m
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

// ContextWithDeadline returns a new context with the given deadline. If the given time
// is in the past, the returned context is already done.
//
// The deadline is set using the clock in the given context.  If there is no
// clock in the context the system clock is used and the result is the same as
// calling context.WithDeadline.
//
// If the context contains a mock clock, the deadline will expire when that
// mock clock is advanced to the deadline or later.
func ContextWithDeadline(ctx context.Context, t time.Time) (context.Context, context.CancelFunc) {
	return ClockFromContext(ctx).ContextWithDeadline(ctx, t)
}

// ContextWithDeadlineCause returns a new context with the given deadline and cause.
// If the given time is in the past, the returned context is already done.
//
// The cause is used to set the context error.
//
// The deadline is set using the clock in the given context.  If there is no
// clock in the context the system clock is used and the result is the same as
// calling context.WithDeadlineCause.
//
// If the context contains a mock clock, the deadline will expire when that
// mock clock is advanced to the deadline or later.
func ContextWithDeadlineCause(ctx context.Context, t time.Time, cause error) (context.Context, context.CancelFunc) {
	return ClockFromContext(ctx).ContextWithDeadlineCause(ctx, t, cause)
}

// ContextWithTimeout returns a new context with the given timeout.
// If the given duration is zero or negative, the returned context is already done.
//
// The timeout is set using the clock in the given context.  If there is no
// clock in the context the system clock is used and the result is the same as
// calling context.WithTimeout.
//
// If the context contains a mock clock, the timeout will expire when that
// mock clock is advanced by at least the given duration from its current time.
func ContextWithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return ClockFromContext(ctx).ContextWithTimeout(ctx, d)
}

// ContextWithTimeoutCause returns a new context with the given timeout and cause.
// If the given duration is zero or negative, the returned context is already done.
//
// The cause is used to set the context error.
//
// The timeout is set using the clock in the given context.  If there is no
// clock in the context the system clock is used and the result is the same as
// calling context.WithTimeoutCause.
//
// If the context contains a mock clock, the timeout will expire when that
// mock clock is advanced by at least the given duration from its current time.
// The cause is used to set the context error.
func ContextWithTimeoutCause(ctx context.Context, d time.Duration, cause error) (context.Context, context.CancelFunc) {
	return ClockFromContext(ctx).ContextWithTimeoutCause(ctx, d, cause)
}
