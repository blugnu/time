package internal_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	stdlib "time"

	. "github.com/blugnu/test"
	"github.com/blugnu/test/opt"
	"github.com/blugnu/time"
	"github.com/blugnu/time/internal"
)

// Ensure that the clock's After channel sends at the correct stdlib.
func TestSystemClock_After(t *testing.T) {
	With(t)

	dur := time.Dur(func() {
		<-internal.SystemClockInstance.After(20 * stdlib.Millisecond)
	})

	if dur < 20*stdlib.Millisecond || dur > 40*stdlib.Millisecond {
		Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock's AfterFunc executes at the correct stdlib.
func TestSystemClock_AfterFunc(t *testing.T) {
	With(t)

	var (
		ok  bool
		wg  sync.WaitGroup
		sys = internal.SystemClockInstance
	)

	wg.Add(1)
	dur := time.Dur(func() {
		sys.AfterFunc(20*stdlib.Millisecond, func() {
			ok = true
			wg.Done()
		})
		wg.Wait()
	})

	if dur < 20*stdlib.Millisecond || dur > 40*stdlib.Millisecond {
		Fatalf("Bad duration: %s", dur)
	}
	if !ok {
		Fatal("Function did not run")
	}
}

// Ensure that the clock's time matches the standary library.
func TestSystemClock_Now(t *testing.T) {
	With(t)

	sys := internal.SystemClockInstance

	a := stdlib.Now().Round(stdlib.Second)
	b := sys.Now().Round(stdlib.Second)
	if !a.Equal(b) {
		t.Errorf("not equal: %s != %s", a, b)
	}
}

func TestSystemClock_Since(t *testing.T) {
	With(t)

	var (
		clock = internal.SystemClockInstance
		start = clock.Now()
		dur   time.Duration
	)
	stdlib.Sleep(10 * stdlib.Millisecond)
	dur = clock.Since(start)

	Expect(dur).ToNot(BeLessThan(10 * stdlib.Millisecond))
}

// Ensure that the clock sleeps for the appropriate amount of stdlib.
func TestSystemClock_Sleep(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	dur := time.Dur(func() {
		clock.Sleep(20 * stdlib.Millisecond)
	})

	Expect(dur).ToNot(BeLessThan(20 * stdlib.Millisecond))
	Expect(dur).ToNot(BeGreaterThan(40 * stdlib.Millisecond))
}

// Ensure that the clock ticks correctly.
func TestSystemClock_Tick(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	dur := time.Dur(func() {
		c := clock.Tick(20 * stdlib.Millisecond)
		<-c
		<-c
	})

	// Expect(dur).To(BeBetween(MinMax(20, 50, stdlib.Millisecond)...), opt.Inclusive)
	Expect(dur).ToNot(BeLessThan(20 * stdlib.Millisecond))
	Expect(dur).ToNot(BeGreaterThan(50 * stdlib.Millisecond))
}

// Ensure that the clock's ticker ticks correctly.
func TestSystemClock_Ticker(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	dur := time.Dur(func() {
		ticker := clock.Tick(50 * stdlib.Millisecond)
		<-ticker
		<-ticker
	})

	Expect(dur).To(BeBetween(100*stdlib.Millisecond).And(200*time.Millisecond), opt.IntervalClosed)
}

func TestSystemClock_Until(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	start := stdlib.Now()
	dur := clock.Until(start.Add(10 * stdlib.Millisecond))

	Expect(dur).To(BeLessThan(10 * stdlib.Millisecond).OrEqual())
}

// Ensure that the clock's ticker can stop correctly.
func TestSystemClock_Ticker_Stop(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	ticker := clock.NewTicker(20 * stdlib.Millisecond)
	<-ticker.C
	ticker.Stop()
	select {
	case <-ticker.C:
		Fatal("unexpected send")
	case <-stdlib.After(30 * stdlib.Millisecond):
	}
}

// Ensure that the clock's ticker can reset correctly.
func TestSystemClock_Ticker_Reset(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	ticker := clock.NewTicker(20 * stdlib.Millisecond)
	dur := time.Dur(func() {
		<-ticker.C // ~20ms
		ticker.Reset(5 * stdlib.Millisecond)
		<-ticker.C // + ~5ms (=> ~25ms)
		ticker.Stop()
	})

	Expect(dur).ToNot(BeLessThan(25 * stdlib.Millisecond))
	Expect(dur).ToNot(BeGreaterThan(30 * stdlib.Millisecond))
}

// Ensure that the clock's ticker can stop and then be reset correctly.
func TestSystemClock_Ticker_StopThenReset(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	ticker := clock.NewTicker(20 * stdlib.Millisecond)
	dur := time.Dur(func() {
		<-ticker.C // ~20ms
		ticker.Stop()
		select {
		case <-ticker.C:
			Fatal("unexpected send")
		case <-stdlib.After(30 * stdlib.Millisecond): // + ~30ms while we wait to ensure ticker is stopped (=> 50ms)
		}
		ticker.Reset(5 * stdlib.Millisecond)
		<-ticker.C // + ~5ms (=> ~55ms)
		ticker.Stop()
	})

	Expect(dur).ToNot(BeLessThan(55 * stdlib.Millisecond))
	Expect(dur).ToNot(BeGreaterThan(60 * stdlib.Millisecond))
}

// Ensure that the clock's timer waits correctly.
func TestSystemClock_Timer(t *testing.T) {
	With(t)

	clock := internal.SystemClockInstance

	timer := clock.NewTimer(20 * stdlib.Millisecond)
	dur := time.Dur(func() {
		<-timer.C
	})

	Expect(dur).ToNot(BeLessThan(20 * stdlib.Millisecond))
	Expect(dur).ToNot(BeGreaterThan(40 * stdlib.Millisecond))

	Expect(timer.Stop(), "timer already fired").To(BeFalse())
}

// Tests that a context created using ContextWithTimeoutCause that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithDeadlineCause(t *testing.T) {
	With(t)

	var (
		ctx       = context.Background()
		clock     = internal.SystemClockInstance
		cause     = errors.New("test cause")
		cancelled atomic.Bool
		listener  internal.WaitFuncs
	)

	// act: create a context with a deadline with cause using the system clock
	//      and start a goroutine that waits for it to be cancelled
	ctx, _ = clock.ContextWithDeadlineCause(ctx, clock.Now().Add(5*time.Millisecond), cause)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	// assert: that the context is cancelled and the error is DeadlineExceeded
	Expect(cancelled.Load(), "context cancelled").To(BeTrue())
	Expect(ctx.Err(), "error").Is(context.DeadlineExceeded)
	Expect(context.Cause(ctx), "cause").Is(cause)
}

// Tests that a context created using ContextWithDeadline that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithDeadline(t *testing.T) {
	With(t)

	// arrange
	var (
		ctx       = context.Background()
		clock     = internal.SystemClockInstance
		cancelled atomic.Bool
		listener  internal.WaitFuncs
	)

	// act: create a context with a deadline using the system clock and
	//      start a goroutine that waits for it to be cancelled
	ctx, _ = clock.ContextWithDeadline(ctx, clock.Now().Add(5*stdlib.Millisecond))
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	// assert: that the context is cancelled and the error is DeadlineExceeded
	Expect(cancelled.Load(), "context cancelled").To(BeTrue())
	Expect(ctx.Err()).Is(context.DeadlineExceeded)
}

// Tests that a context created using ContextWithTimeoutCause that is not mocked
// is cancelled when deadline is reached.
func Test_ContextWithTimeoutCause(t *testing.T) {
	With(t)

	// arrange
	var (
		ctx       = context.Background()
		clock     = internal.SystemClockInstance
		cancelled atomic.Bool
		cause     = errors.New("test cause")
		listener  internal.WaitFuncs
	)

	// act: create a context with a timeout with cause using the system clock and
	//      start a goroutine that waits for it to be cancelled
	ctx, _ = clock.ContextWithTimeoutCause(ctx, 5*time.Millisecond, cause)
	listener.Go(func() {
		<-ctx.Done()
		cancelled.Store(true)
	})
	listener.Wait()

	// assert: that the context is cancelled and the error is DeadlineExceeded
	Expect(cancelled.Load(), "context cancelled").To(BeTrue())
	Expect(ctx.Err(), "error").Is(context.DeadlineExceeded)
	Expect(context.Cause(ctx), "cause").Is(cause)
}

// Tests that a context created using ContextWithTimeoutCause with a parent deadline
// occurring sooner is cancelled when the parent deadline is reached.
func Test_ContextWithTimeoutCause_ParentDeadlineIsEarlier(t *testing.T) {
	With(t)

	// arrange
	var (
		ctx      = context.Background()
		parent   context.Context
		child    context.Context
		clock    = internal.SystemClockInstance
		cause    = errors.New("test cause")
		listener internal.WaitFuncs
		elapsed  stdlib.Duration
	)
	parent, _ = clock.ContextWithTimeout(ctx, 5*time.Millisecond)
	child, _ = clock.ContextWithTimeoutCause(parent, 100*stdlib.Millisecond, cause)

	listener.Go(func() {
		elapsed = time.Dur(func() {
			<-child.Done()
		})
	})
	listener.Wait()

	Expect(elapsed).To(BeLessThan(8 * stdlib.Millisecond))
}
