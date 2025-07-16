package time_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	. "github.com/blugnu/test"

	"github.com/blugnu/time"
	"github.com/blugnu/time/internal"
)

// Tests that FromContext returns the clock present in the context.
func Test_FromContext(t *testing.T) {
	With(t)

	ctx := context.Background()

	Run(Test("with no clock in context", func() {
		// act
		clock := time.FromContext(ctx)

		// assert
		_ = RequireType[internal.SystemClock](clock)
	}))

	Run(Test("with clock in context", func() {
		// arrange
		mock := internal.NewMockClock()
		ctx := time.ContextWithClock(ctx, mock)

		// act
		clock := time.FromContext(ctx)

		// assert
		_ = RequireType[*internal.MockClock](clock)
	}))
}

// Tests that ContextWithClock returns a new context with the given clock.
func Test_ContextWithClock(t *testing.T) {
	With(t)

	// arrange
	var (
		mock = internal.NewMockClock()
		ctx  = context.Background()
	)

	// act
	ctx = time.ContextWithClock(ctx, mock)

	// assert
	Expect(time.FromContext(ctx)).Is(mock)
}

// Tests that ContextWithClock panics if the parent context already has a clock.
func Test_ContextWithClock_ParentContextHasClock(t *testing.T) {
	With(t)

	// arrange/assert
	ctx := context.Background()
	ctx = time.ContextWithClock(ctx, time.SystemClock())
	defer Expect(Panic(time.ErrClockAlreadyExists)).DidOccur()

	// act
	_ = time.ContextWithClock(ctx, time.SystemClock())
}

// Tests that ContextWithClock returns a context with the system clock if nil is
// given as the clock.
func Test_ContextWithClock_NilClockGiven(t *testing.T) {
	With(t)

	// arrange
	ctx := context.Background()

	// act
	ctx = time.ContextWithClock(ctx, nil)

	// assert
	clock := time.FromContext(ctx)
	RequireType[internal.SystemClock](clock)
}

// Tests that ContextWithMockClock returns a new context with a mock clock.
func Test_ContextWithMockClock(t *testing.T) {
	With(t)

	// arrange
	ctx := context.Background()

	// act
	ctx, mock := time.ContextWithMockClock(ctx)

	// assert
	Expect(time.FromContext(ctx)).Is(mock)
}

// Tests that ContextWithMockClock panics if the parent context already has a clock.
func Test_ContextWithMockClock_ParentContextHasClock(t *testing.T) {
	With(t)

	// arrange/assert
	ctx := time.ContextWithClock(context.Background(), time.SystemClock())
	defer Expect(Panic(time.ErrClockAlreadyExists)).DidOccur()

	// act
	_, _ = time.ContextWithMockClock(ctx)
}

// Tests that a context created using ContextWithDeadline that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithDeadline(t *testing.T) {
	With(t)

	// arrange
	var (
		ctx       = context.Background()
		cancelled atomic.Bool
		listener  internal.WaitFuncs
	)

	// act: create a context with a deadline using the system clock
	//      and start a goroutine that waits for it to be cancelled
	ctx, _ = time.ContextWithDeadline(ctx, time.SystemClock().Now().Add(5*time.Millisecond))
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	Expect(cancelled.Load(), "context cancelled").To(BeTrue())
	Expect(ctx.Err()).Is(context.DeadlineExceeded)
	Expect(context.Cause(ctx)).Is(context.DeadlineExceeded)
}

// Tests that a context created using ContextWithTimeoutCause that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithDeadlineCause(t *testing.T) {
	With(t)

	// arrange
	var (
		cause     = errors.New("test cause")
		ctx       = context.Background()
		cancelled atomic.Bool
		listener  internal.WaitFuncs
	)

	// act: create a context with a deadline with cause using the system clock
	//      and start a goroutine that waits for it to be cancelled
	ctx, _ = time.ContextWithDeadlineCause(ctx, time.SystemClock().Now().Add(5*time.Millisecond), cause)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	// assert: that the context is cancelled and the error is DeadlineExceeded with
	//         the expected cause
	Expect(cancelled.Load(), "context cancelled").To(BeTrue())
	Expect(ctx.Err()).Is(context.DeadlineExceeded)
	Expect(context.Cause(ctx)).Is(cause)
}

// Tests that a context created using ContextWithTimeout that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithTimeout(t *testing.T) {
	With(t)

	// arrange
	var (
		ctx       = context.Background()
		cancelled atomic.Bool
		listener  internal.WaitFuncs
	)

	// act: create a context with a timeout using the system clock
	//      and start a goroutine that waits for it to be cancelled
	ctx, _ = time.ContextWithTimeout(ctx, 5*time.Millisecond)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	Expect(cancelled.Load(), "context cancelled").To(BeTrue())
	Expect(ctx.Err(), "error").Is(context.DeadlineExceeded)
	Expect(context.Cause(ctx), "cause").Is(context.DeadlineExceeded)
}

// Tests that a context created using ContextWithTimeoutCause that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithTimeoutCause(t *testing.T) {
	With(t)

	// arrange
	var (
		cause     = errors.New("test cause")
		ctx       = context.Background()
		cancelled atomic.Bool
		listener  internal.WaitFuncs
	)

	// act: create a context with a timeout with cause using the system clock
	//      and start a goroutine that waits for it to be cancelled
	ctx, _ = time.ContextWithTimeoutCause(ctx, 5*time.Millisecond, cause)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	// assert: that the context is cancelled and the error is DeadlineExceeded with
	//         the expected cause
	Expect(cancelled.Load(), "context cancelled").To(BeTrue())
	Expect(ctx.Err(), "error").Is(context.DeadlineExceeded)
	Expect(context.Cause(ctx), "cause").Is(cause)
}
