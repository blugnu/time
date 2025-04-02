package time

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blugnu/test"
)

// Tests that ClockFromContext returns the clock present in the context.
func Test_ClockFromContext(t *testing.T) {
	// arrange
	mock := NewMockClock()
	ctx := ContextWithClock(context.Background(), mock)

	// act
	clock := ClockFromContext(ctx)

	// assert
	test.Value(t, clock).Equals(mock)
}

// Tests that ClockFromContext returns the system clock if no clock is
// present in the context.
func Test_ClockFromContext_NoClockPresent(t *testing.T) {
	// act
	clock := ClockFromContext(context.Background())

	// assert
	test.IsType[systemClock](t, clock)
}

// Tests that ContextWithClock returns a new context with the given clock.
func Test_ContextWithClock(t *testing.T) {
	// arrange
	mock := NewMockClock()
	bg := context.Background()

	// act
	ctx := ContextWithClock(bg, mock)

	// assert
	test.Value(t, ctx).DoesNotEqual(bg)
	test.IsTrue(t, ctx.Value(clockKey) == mock)
}

// Tests that ContextWithClock panics if the parent context already has a clock.
func Test_ContextWithClock_ParentContextHasClock(t *testing.T) {
	// arrange
	parent := ContextWithClock(context.Background(), SystemClock())
	defer test.ExpectPanic(ErrClockAlreadyExists).Assert(t)

	// act
	_ = ContextWithClock(parent, SystemClock())
}

// Tests that ContextWithClock returns a context with the system clock if nil is
// given as the clock.
func Test_ContextWithClock_NilClockGiven(t *testing.T) {
	// act
	ctx := ContextWithClock(context.Background(), nil)

	// assert
	clock := ClockFromContext(ctx)
	test.IsType[systemClock](t, clock)
}

// Tests that ContextWithMockClock returns a new context with a mock clock.
func Test_ContextWithMockClock(t *testing.T) {
	// arrange
	bg := context.Background()

	// act
	ctx, mock := ContextWithMockClock(bg)

	// assert
	test.Value(t, ctx, "returns new context").DoesNotEqual(bg)
	test.IsNotNil(t, mock, "mock clock")
}

// Tests that ContextWithMockClock panics if the parent context already has a clock.
func Test_ContextWithMockClock_ParentContextHasClock(t *testing.T) {
	// arrange
	parent := ContextWithClock(context.Background(), SystemClock())
	defer test.ExpectPanic(ErrClockAlreadyExists).Assert(t)

	// act
	_, _ = ContextWithMockClock(parent)
}

// Tests that a context created using ContextWithTimeoutCause that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithDeadlineCause(t *testing.T) {
	ctx := ContextWithClock(context.Background(), SystemClock())
	ctx, _ = ContextWithDeadlineCause(ctx, SystemClock().Now().Add(2*time.Millisecond), nil)

	var (
		cancelled atomic.Bool
		listener  WaitFuncs
	)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	test.IsTrue(t, cancelled.Load(), "context cancelled")
	test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
}

// Tests that a context created using ContextWithDeadline that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithDeadline(t *testing.T) {
	ctx := ContextWithClock(context.Background(), SystemClock())
	ctx, _ = ContextWithDeadline(ctx, SystemClock().Now().Add(2*time.Millisecond))

	var (
		cancelled atomic.Bool
		listener  WaitFuncs
	)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	time.Sleep(3 * time.Millisecond)
	listener.Wait()

	test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
	test.IsTrue(t, cancelled.Load(), "context cancelled when deadline reached")
}

// Tests that a context created using ContextWithTimeoutCause that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithTimeoutCause(t *testing.T) {
	ctx := ContextWithClock(context.Background(), SystemClock())
	ctx, _ = ContextWithTimeoutCause(ctx, 2*time.Millisecond, nil)

	var (
		cancelled atomic.Bool
		listener  WaitFuncs
	)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	time.Sleep(3 * time.Millisecond)
	listener.Wait()

	test.IsTrue(t, cancelled.Load(), "context cancelled when deadline reached")
}

// Tests that a context created using ContextWithTimeoutCause with a parent deadline
// occurring sooner is cancelled when the parent deadline is reached.
func Test_ContextWithTimeoutCause_ParentDeadlineIsEarlier(t *testing.T) {
	ctx := ContextWithClock(context.Background(), SystemClock())
	parent, _ := ContextWithTimeout(ctx, 2*time.Millisecond)
	child, _ := ContextWithTimeoutCause(parent, 100*time.Millisecond, nil)

	var (
		listener WaitFuncs
		dur      time.Duration
	)
	listener.Go(func() {
		start := time.Now()
		<-child.Done()
		dur = time.Since(start)
	})
	time.Sleep(3 * time.Millisecond)
	listener.Wait()

	test.IsTrue(t, dur < 5*time.Millisecond, "cancelled when parent deadline reached")
}

// Tests the string representation of a mocked context.
func Test_Mocked_Context_String(t *testing.T) {
	ctx, mock := ContextWithMockClock(context.Background())
	ctx, cancel := ContextWithDeadline(ctx, mock.Now().Add(time.Second))
	defer cancel()

	// act
	str := ctx.(fmt.Stringer).String()

	// assert
	test.Value(t, str).Equals("mock: context.WithDeadline: 1s: 1970-01-01 00:00:01 +0000 UTC")

	// arrange: advance the clock to reduce the duration to the deadline
	mock.AdvanceBy(500 * time.Millisecond)

	// act
	str = ctx.(fmt.Stringer).String()

	// assert: the string representation is updated with the reduced duration
	test.Value(t, str).Equals("mock: context.WithDeadline: 500ms: 1970-01-01 00:00:01 +0000 UTC")
}

// Tests that a mocked ContextWithDeadline is cancelled when the mock clock is
// advanced to the deadline.
func Test_Mocked_ContextWithDeadline(t *testing.T) {
	ctx, m := ContextWithMockClock(context.Background())
	ctx, _ = ContextWithDeadline(ctx, m.Now().Add(time.Second))
	m.AdvanceBy(time.Second)
	select {
	case <-ctx.Done():
		test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
	default:
		t.Error("context is not cancelled when deadline exceeded")
	}
}

// Tests that a mocked ContextWithDeadlineCause wraps the cause error
// and is cancelled when the mock clock is advanced to the deadline.
func Test_Mocked_ContextWithDeadlineCause(t *testing.T) {
	cause := errors.New("cause")
	ctx, m := ContextWithMockClock(context.Background())
	ctx, _ = ContextWithDeadlineCause(ctx, m.Now().Add(time.Second), cause)
	m.AdvanceBy(time.Second)
	select {
	case <-ctx.Done():
		test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
		test.Error(t, ctx.Err()).Is(cause)
		test.String(t, ctx.Err().Error()).Equals("context deadline exceeded: cause")
	default:
		t.Error("context was not cancelled")
	}
}

// Tests that a mocked ContextWithDeadline does nothing when the deadline
// is later than a deadline in the parent context.
func Test_Mocked_ContextWithDeadline_LaterThanParent(t *testing.T) {
	m := NewMockClock()
	ctx := ContextWithClock(context.Background(), m)
	ctx, _ = ContextWithDeadline(ctx, m.Now().Add(time.Second))
	ctx, _ = ContextWithDeadline(ctx, m.Now().Add(10*time.Second))
	m.AdvanceBy(time.Second)
	select {
	case <-ctx.Done():
		test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
	default:
		t.Error("context was not cancelled")
	}
}

// Tests that the cancel func returned by a mocked ContextWithDeadline cancels
// the context correctly without needing to advance the clock.
func Test_Mocked_ContextWithDeadline_Cancel(t *testing.T) {
	// arrange
	dur := 10 * time.Millisecond
	ctx, mock := ContextWithMockClock(context.Background())
	ctx, cancel := ContextWithDeadline(ctx, mock.Now().Add(dur))

	// act
	cancel()

	// assert
	select {
	case <-ctx.Done():
		test.Error(t, ctx.Err()).Is(context.Canceled)
	case <-time.After(dur):
		t.Error("context was not cancelled")
	}
}

// Tests that a mocked ContextWithDeadline cancels a context if the parent context
// is cancelled.
func Test_Mocked_ContextWithDeadline_ParentCancelled(t *testing.T) {
	// arrange
	ctx, mock := ContextWithMockClock(context.Background())
	parent, cancelParent := context.WithCancel(ctx)
	child, _ := ContextWithDeadline(parent, mock.Now().Add(time.Second))

	// act: cancel the parent context
	cancelParent()
	select {
	case <-child.Done():
		test.Error(t, child.Err()).Is(context.Canceled)
	case <-time.After(time.Second):
		t.Error("child context was not cancelled")
	}
}

// Tests that a mocked ContextWithDeadline does not cancel parent when cancelled before
// the parent.
func Test_Mocked_ContextWithDeadline_ChildCancelled(t *testing.T) {
	// arrange
	bg, mock := ContextWithMockClock(context.Background())
	parent, cancelParent := ContextWithDeadline(bg, mock.Now().Add(10*time.Millisecond))
	defer cancelParent()

	child, cancelChild := ContextWithDeadline(parent, mock.Now().Add(10*time.Millisecond))

	// act: cancel the child context
	cancelChild()

	// assert: the child context is cancelled
	select {
	case <-child.Done():
		test.Error(t, child.Err()).Is(context.Canceled)
	default:
		t.Error("child context was not cancelled")
	}

	// act: advance the clock to the deadline of the parent
	mock.AdvanceBy(10 * time.Millisecond)

	// assert: the parent context is now expired
	select {
	case <-parent.Done():
		test.Error(t, parent.Err()).Is(context.DeadlineExceeded)
	default:
		t.Error("parent context was not cancelled")
	}
}

// Tests that a mock ContextWithDeadline with a deadline that has already passed
// is cancelled immediately and returns a no-op cancel function.
func Test_Mocked_ContextWithDeadline_DeadlineAlreadyPassed(t *testing.T) {
	// arrange
	ctx, clock := ContextWithMockClock(context.Background())

	// act: create a context with a deadline in the past
	ctx, cancel := ContextWithDeadline(ctx, clock.Now().Add(-time.Second))

	// assert: the context is cancelled immediately
	select {
	case <-ctx.Done():
		test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
	case <-time.After(time.Millisecond):
		t.Error("context was not immediately cancelled")
	}

	// act: cancel the context
	cancel()

	// assert: cancellation did not change the error
	test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
}

// Tests that a context created using ContextWithTimeout is cancelled when
// deadline is reached.
func Test_Mocked_ContextWithTimeout(t *testing.T) {
	ctx, clock := ContextWithMockClock(context.Background())
	ctx, _ = ContextWithTimeout(ctx, time.Second)
	clock.AdvanceBy(time.Second)
	select {
	case <-ctx.Done():
		test.Error(t, ctx.Err()).Is(context.DeadlineExceeded)
	default:
		t.Error("context was not cancelled")
	}
}
