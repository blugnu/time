package internal_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/blugnu/test"
	"github.com/blugnu/time/internal"
)

func Fatal(msg string) {
	T().Helper()
	T().Fatal(msg)
}

func Fatalf(msg string, args ...any) {
	T().Helper()
	T().Fatalf(msg, args...)
}

// Test that a ticker established by After sends at the correct time.
func TestMock_After(t *testing.T) {
	With(t)

	var (
		clock    = internal.NewMockClock()
		ticked   atomic.Bool
		listener internal.WaitFuncs
	)

	// Create a channel to execute after 10 mock seconds.
	ch := clock.After(10 * time.Second)
	listener.Go(func() {
		<-ch
		ticked.Store(true)
	})

	// Move clock forward to just before the time.
	clock.AdvanceBy(9 * time.Second)
	Expect(ticked.Load(), "fired early").To(BeFalse())

	// Move clock forward to the after channel's time.
	clock.AdvanceBy(1 * time.Second)
	listener.Wait()
	Expect(ticked.Load(), "fired on time").To(BeTrue())
}

// Ensure that the mock's After channel doesn't block on write.
func TestMock_UnusedAfter(t *testing.T) {
	With(t)

	mock := internal.NewMockClock()
	mock.After(1 * time.Millisecond)

	done := internal.Waitable(func() {
		mock.AdvanceBy(1 * time.Second)
	})

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		Fatal("mock.AdvanceBy hung")
	}
}

// Ensure that the mock's AfterFunc executes at the correct time.
func TestMock_AfterFunc(t *testing.T) {
	With(t)

	var ticked atomic.Bool
	clock := internal.NewMockClock()

	// Execute function after duration.
	clock.AfterFunc(10*time.Second, func() {
		ticked.Store(true)
	})

	// Move clock forward to just before the time.
	clock.AdvanceBy(9 * time.Second)
	Expect(ticked.Load(), "fired early").To(BeFalse())

	// Move clock forward to the after channel's time.
	clock.AdvanceBy(1 * time.Second)
	Expect(ticked.Load(), "fired on time").To(BeTrue())
}

// Ensure that the mock's AfterFunc doesn't execute if stopped.
func TestMock_AfterFunc_Stop(t *testing.T) {
	With(t)

	// Execute function after duration.
	clock := internal.NewMockClock()
	timer := clock.AfterFunc(10*time.Second, func() {
		Fatal("unexpected function execution")
	})

	// Stop timer & move clock forward.
	timer.Stop()
	clock.AdvanceBy(10 * time.Second)
}

// Ensure that the mock's current time can be changed.
func TestMock_Now(t *testing.T) {
	With(t)

	clock := internal.NewMockClock()
	if now := clock.Now(); !now.Equal(time.Unix(0, 0)) {
		Fatalf("expected epoch, got: %v", now)
	}

	// Add 10 seconds and check the time.
	clock.AdvanceBy(10 * time.Second)
	if now := clock.Now(); !now.Equal(time.Unix(10, 0)) {
		Fatalf("expected epoch, got: %v", now)
	}
}

// Test that IsRunning returns the state of the clock.
func TestMock_IsRunning(t *testing.T) {
	With(t)

	// arrange: create a clock in default (stopped) state
	clock := internal.NewMockClock()

	// assert: that the clock is not running
	Expect(clock.IsRunning()).To(BeFalse())

	// act/assert: start the clock and check that it is running
	clock.Start()
	Expect(clock.IsRunning()).To(BeTrue())
}

// Tests that Since returns the duration since the clock was created.
func TestMock_Since(t *testing.T) {
	With(t)

	Run(Test("advancing a frozen clock", func() {
		clock := internal.NewMockClock()

		beginning := clock.Now()
		clock.AdvanceBy(500 * time.Second)

		Expect(clock.Since(beginning).Seconds()).To(Equal[float64](500))
	}))

	Run(Test("with running clock", func() {
		clock := internal.NewMockClock(internal.StartRunning())
		time.Sleep(25 * time.Millisecond)

		Expect(clock.Since(time.Unix(0, 0)).Milliseconds() >= 25).To(BeTrue())
	}))
}

func TestMock_Until(t *testing.T) {
	With(t)

	clock := internal.NewMockClock()

	end := clock.Now().Add(500 * time.Second)
	Expect(clock.Until(end)).To(Equal(500 * time.Second))

	clock.AdvanceBy(100 * time.Second)
	Expect(clock.Until(end)).To(Equal(400 * time.Second))
}

// Test that Sleep respects the passage of mocked time for a stopped clock.
func TestMock_Sleep_StoppedClock(t *testing.T) {
	With(t)

	// arrange: start a goroutine that sleeps for 10 seconds
	var (
		clock = internal.NewMockClock()
		ok    atomic.Bool
	)
	go func() {
		clock.Sleep(10 * time.Second)
		ok.Store(true)
	}()

	// act/assert: after 9 mock seconds, the goroutine should still be sleeping
	clock.AdvanceBy(9 * time.Second)
	Expect(ok.Load(), "woke early").To(BeFalse())

	// act/assert: after 1 more second, the goroutine should have awoken
	clock.AdvanceBy(1 * time.Second)
	Expect(ok.Load(), "woke when expected").To(BeTrue())
}

// Tests that negative Sleep returns immediately when clock is running.
func TestMock_Sleep_Negative_RunningClock(t *testing.T) {
	With(t)

	// arrange: create a clock in running state and sleep for -1ms
	var (
		clock = internal.NewMockClock(internal.StartRunning())
		start = time.Now()
		dur   time.Duration
	)
	internal.WaitFor(func() {
		clock.Sleep(-1 * time.Millisecond)
		dur = time.Since(start)
	})

	// act/assert: the clock should not have advanced
	Expect(dur < 1*time.Millisecond).To(BeTrue())
}

// Tests that negative Sleep returns immediately when clock is stopped.
func TestMock_Sleep_Negative_StoppedClock(t *testing.T) {
	With(t)

	// arrange: create a clock in stopped state and sleep for -1ms
	var (
		clock = internal.NewMockClock()
	)
	internal.WaitFor(func() {
		clock.Sleep(-1 * time.Millisecond)
	})

	// act/assert: the clock should not have advanced
	Expect(clock.SinceCreated() == 0).To(BeTrue())
}

// Tests that Sleep respects the passage of elapsed time for a running clock.
func TestMock_Sleep_RunningClock(t *testing.T) {
	With(t)

	// arrange: create a clock in running state and sleep for 10ms
	clock := internal.NewMockClock(internal.StartRunning())
	time.Sleep(10 * time.Millisecond)

	// act/assert: the clock should have advanced by at least 10ms
	Expect(clock.SinceCreated() >= 10*time.Millisecond).To(BeTrue())
}

// Tests that a zero Tick duration returns a nil channel.
func TestMock_Tick_Zero(t *testing.T) {
	With(t)

	// arrange: create a clock and a channel to receive ticks
	clock := internal.NewMockClock()
	tick := clock.Tick(0)

	// act/assert: the tick channel should be nil
	Expect(tick).IsNil()
}

// Tests that a negative Tick duration returns a nil channel.
func TestMock_Tick_Negative(t *testing.T) {
	With(t)

	// arrange: create a clock and a channel to receive ticks
	clock := internal.NewMockClock()
	tick := clock.Tick(-1 * time.Second)

	// act/assert: the tick channel should be nil
	Expect(tick).IsNil()
}

// Tests that Start resumes a stopped clock.
func TestMock_Start(t *testing.T) {
	With(t)

	// arrange: create a default (stopped) clock
	clock := internal.NewMockClock()

	// act/assert: sleep for 10ms and check that the clock has not advanced
	time.Sleep(10 * time.Millisecond)
	Expect(clock.SinceCreated() == 0).To(BeTrue())

	// act: start the clock and sleep for another 10ms
	clock.Start()
	time.Sleep(10 * time.Millisecond)

	// assert: the mock time should reflect the passing of at least 10ms in real time
	Expect(clock.SinceCreated() >= 10*time.Millisecond).To(BeTrue())
}

// Tests that Start panics if the clock is already running.
func TestMock_Start_Running(t *testing.T) {
	With(t)

	// arrange/assert: create a default (running) clock
	clock := internal.NewMockClock(internal.StartRunning())
	defer Expect(Panic(internal.ErrClockIsRunning)).DidOccur()

	// act: attempt to start the clock (again)
	clock.Start()
}

// Tests that Stop pauses a running clock.
func TestMock_Stop(t *testing.T) {
	With(t)

	// arrange: create a default (running) clock
	clock := internal.NewMockClock(internal.StartRunning())

	// assert: verify that the clock is running;
	//  sleep for 10ms and ensure that mock time reflects the elapsed time
	time.Sleep(10 * time.Millisecond)
	Expect(clock.SinceCreated() >= 10*time.Millisecond).To(BeTrue())

	// act/assert: stop the clock, record the mock time then sleep for 10ms
	//  and verify that the mock time has not advanced
	clock.Stop()
	stoppedAt := clock.Now()
	time.Sleep(10 * time.Millisecond)
	Expect(clock.Since(stoppedAt) == 0).To(BeTrue())
}

// Tests that a channel established by Tick sends at the correct time.
func TestMock_Tick(t *testing.T) {
	With(t)

	// arrange: start a ticker to fire every 10 seconds and count the ticks
	var (
		ticks atomic.Uint32
		clock = internal.NewMockClock()
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	tick := clock.Tick(10 * time.Second)
	go func() {
		for {
			select {
			case <-tick:
				ticks.Add(1)
			case <-done:
				return
			}
		}
	}()

	// act/assert: there should be no ticks until the clock is advanced to the
	// first tick time
	clock.AdvanceBy(9 * time.Second)
	Expect(ticks.Load()).To(Equal[uint32](0))

	// act/assert: after 1 more second, the first tick should have fired
	clock.AdvanceBy(1 * time.Second)
	Expect(ticks.Load()).To(Equal[uint32](1))

	// act/assert: after 20 more seconds there should have been 2 further ticks
	clock.AdvanceBy(20 * time.Second)
	Expect(ticks.Load()).To(Equal[uint32](3))
}

// Tests that a Ticker channel sends at the correct time.
func TestMock_Ticker(t *testing.T) {
	With(t)

	var (
		cnt   atomic.Uint32
		clock = internal.NewMockClock()
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	// start a go routine that will count the ticks
	ticker := clock.NewTicker(1 * time.Microsecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				cnt.Add(1)
			case <-done:
				return
			}
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(10 * time.Microsecond)
	Expect(cnt.Load()).To(Equal[uint32](10))
}

// Tests that a Ticker with zero duration fires immediately.
func TestMock_Ticker_Zero(t *testing.T) {
	With(t)

	// arrange: create a clock and a channel to receive ticks
	clock := internal.NewMockClock()
	ticker := clock.NewTicker(0)

	// assert: the tick channel should not be nil
	Expect(ticker.C).IsNotNil()

	// assert: the ticker ticked at the clock creation time
	tm := <-ticker.C
	Expect(clock.Since(tm) == 0).To(BeTrue())
}

// Ensure that the mock's Ticker channel won't block if not read from.
func TestMock_Ticker_Overflow(t *testing.T) {
	With(t)

	clock := internal.NewMockClock()
	ticker := clock.NewTicker(1 * time.Microsecond)
	clock.AdvanceBy(10 * time.Microsecond)
	ticker.Stop()
}

// Ensure that the mock's Ticker can be stopped.
func TestMock_Ticker_Stop(t *testing.T) {
	With(t)

	var (
		cnt   atomic.Uint32
		clock = internal.NewMockClock()
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	// act: start a goroutine that will count the ticks
	ticker := clock.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				cnt.Add(1)
			case <-done:
				return
			}
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(5 * time.Second)
	Expect(cnt.Load()).To(Equal[uint32](5))

	ticker.Stop()

	// Move clock forward again.
	clock.AdvanceBy(5 * time.Second)
	Expect(cnt.Load()).To(Equal[uint32](5))
}

func TestMock_Ticker_Reset(t *testing.T) {
	With(t)

	var (
		cnt   atomic.Uint32
		clock = internal.NewMockClock()
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	ticker := clock.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				cnt.Add(1)
			case <-done:
				return
			}
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(10 * time.Second)
	Expect(cnt.Load()).To(Equal[uint32](2))

	clock.AdvanceBy(4 * time.Second)
	ticker.Reset(5 * time.Second)

	// Advance the remaining second
	clock.AdvanceBy(1 * time.Second)

	Expect(cnt.Load()).To(Equal[uint32](2))

	// Advance the remaining 4 seconds from the previous tick
	clock.AdvanceBy(4 * time.Second)

	Expect(cnt.Load()).To(Equal[uint32](3))
}

func TestMock_Ticker_Stop_Reset(t *testing.T) {
	With(t)

	var (
		cnt   atomic.Uint32
		clock = internal.NewMockClock()
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	ticker := clock.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				cnt.Add(1)
			case <-done:
				return
			}
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(10 * time.Second)
	Expect(cnt.Load()).To(Equal[uint32](2))

	ticker.Stop()

	// Move clock forward again.
	clock.AdvanceBy(5 * time.Second)
	Expect(cnt.Load()).To(Equal[uint32](2))

	ticker.Reset(2 * time.Second)

	// Advance the remaining 2 seconds
	clock.AdvanceBy(2 * time.Second)

	Expect(cnt.Load()).To(Equal[uint32](3))

	// Advance another 2 seconds
	clock.AdvanceBy(2 * time.Second)

	Expect(cnt.Load()).To(Equal[uint32](4))
}

// Ensure that multiple tickers can be used together.
func TestMock_Ticker_Multi(t *testing.T) {
	With(t)

	var (
		cnt   atomic.Uint32
		clock = internal.NewMockClock()
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	go func() {
		ones := clock.NewTicker(1 * time.Microsecond)
		tens := clock.NewTicker(3 * time.Microsecond)

		for {
			select {
			case <-ones.C:
				cnt.Add(1)
			case <-tens.C:
				cnt.Add(10)
			case <-done:
				return
			}
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(10 * time.Microsecond)

	// we're expecting 40 because in 10 microseconds we should have had:
	//
	//  10 ticks from the 1 microsecond ticker = 10
	//  3 ticks from the 3 microsecond ticker = 30
	//
	// 10+30 = 40
	Expect(cnt.Load()).To(Equal[uint32](40))
}

func TestMock_NewTimer_NegativeDuration(t *testing.T) {
	With(t)

	clock := internal.NewMockClock()
	timer := clock.NewTimer(-time.Second)
	select {
	case <-timer.C:
	default:
		Fatal("timer should have fired immediately")
	}
}

func TestMock_Timer_Reset_Zero(t *testing.T) {
	With(t)

	// arrange: create a clock and a timer
	var (
		clock    = internal.NewMockClock()
		timer    = clock.NewTimer(1 * time.Second)
		ticked   atomic.Bool
		dur      time.Duration
		listener internal.WaitFuncs
	)
	listener.Go(func() {
		dur = clock.Since(<-timer.C)
		ticked.Store(true)
	})
	timer.Reset(0)
	listener.Wait()

	// act/assert: the timer should have fired immediately
	Expect(dur).To(Equal[time.Duration](0))
}

func TestMock_ReentrantDeadlock(t *testing.T) {
	With(t)

	var (
		mockedClock = internal.NewMockClock()
		timer20     = mockedClock.NewTimer(20 * time.Second)
		done        = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	go func() {
		select {
		case <-timer20.C:
			panic("timer should not have ticked")
		case <-done:
			return
		}
	}()

	mockedClock.AfterFunc(10*time.Second, func() {
		timer20.Stop()
	})

	mockedClock.AdvanceBy(15 * time.Second)
	mockedClock.AdvanceBy(15 * time.Second)
}

// Test that a running clock advances by the elapsed time
func TestMock_Advance(t *testing.T) {
	With(t)

	// arrange: create a mock clock in running state
	clock := internal.NewMockClock(internal.StartRunning())

	// act: sleep for 100ms then advance the clock
	clock.Sleep(100 * time.Millisecond)
	clock.Update()

	// assert: the clock should have advanced by at least 100ms
	Expect(clock.Since(time.Time{}) >= 100*time.Millisecond).To(BeTrue())
}

// Tests that Advance panics if the clock is not running.
func TestMock_Advance_NotRunning(t *testing.T) {
	With(t)

	// arrange: create a mock clock in stopped state
	clock := internal.NewMockClock()

	// act/assert: attempt to advance the clock (should panic)
	defer Expect(Panic(internal.ErrClockNotRunning)).DidOccur()
	clock.Update()
}

// Tests that AdvanceBy advances the clock by the specified duration.
func TestMock_AdvanceBy(t *testing.T) {
	With(t)

	// arrange: create a mock clock
	clock := internal.NewMockClock()

	// act: advance the clock by 100ms
	clock.AdvanceBy(100 * time.Millisecond)

	// assert: the clock should have advanced by at least 100ms
	Expect(clock.SinceCreated() == 100*time.Millisecond).To(BeTrue())
}

// Tests that AdvanceBy panics if attempting to go back in time.
func TestMock_AdvanceBy_GoingBackInTime(t *testing.T) {
	With(t)

	// arrange: create a mock clock
	clock := internal.NewMockClock()

	// act/assert: attempt to advance the clock back in time
	defer Expect(Panic(internal.ErrNotADelorean)).DidOccur()
	clock.AdvanceBy(-100 * time.Millisecond)
}

// Tests that AdvanceTo advances the clock to the specified time.
func TestMock_AdvanceTo(t *testing.T) {
	With(t)

	// arrange: create a mock clock
	clock := internal.NewMockClock()

	// act: advance the clock to a specific time
	clock.AdvanceTo(time.Unix(100, 0))

	// assert: the clock should have advanced to the specified time
	Expect(clock.Now().Equal(time.Unix(100, 0))).To(BeTrue())
}

// Tests that AdvanceTo panics if attempting to go back in time.
func TestMock_AdvanceTo_GoingBackInTime(t *testing.T) {
	With(t)

	// arrange: create a mock clock
	clock := internal.NewMockClock()

	// act/assert: attempt to advance the clock back in time
	defer Expect(Panic(internal.ErrNotADelorean)).DidOccur()
	clock.AdvanceTo(time.Unix(-100, 0))
}

// Test that many simultaneous timers can be created and that they
// all tick at the correct time.
func TestMock_AfterFuncRace(t *testing.T) {
	With(t)

	var (
		clock  = internal.NewMockClock()
		called atomic.Bool
	)
	defer func() {
		Expect(called.Load(), "func is called").To(BeTrue())
	}()

	funcs := internal.StartFuncs{}
	funcs.OnStart(func() {
		clock.AfterFunc(time.Millisecond, func() {
			called.Store(true)
		})
	})
	funcs.OnStart(func() {
		clock.AdvanceBy(time.Millisecond)
		clock.AdvanceBy(time.Millisecond)
	})

	funcs.Start()
	funcs.Wait()
}

func TestMock_AfterRace(t *testing.T) {
	With(t)

	// arrange: prepare a number of goroutines to setup timers to tick after 1ms.
	// The goroutines will all be started at the same time and will all set their timers
	// from the same base time.

	var (
		mock  = internal.NewMockClock()
		ticks atomic.Int32
		funcs internal.StartFuncs
		n     int32 = 20
	)
	for range n {
		funcs.OnStart(func() {
			<-mock.After(1 * time.Millisecond)
			ticks.Add(1)
		})
	}

	// start the goroutines
	funcs.Start()

	// advance the clock by 1ms
	mock.AdvanceBy(time.Millisecond)

	// wait for all the goroutines to finish
	funcs.Wait()

	// assert: that all the timers ticked
	Expect(ticks.Load(), "ticks").To(Equal(n))
}

func TestMock_DoesNotDropTicks(t *testing.T) {
	With(t)

	// arrange: establish a mock clock with DropsTicks set, create
	// a ticker to tick every 1s and start a goroutine to count the
	// ticks from the ticker
	var (
		cnt   atomic.Uint32
		clock = internal.NewMockClock()
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	ticker := clock.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				cnt.Add(1)
			case <-done:
				return
			}
		}
	}()

	// act: advance the clock by 10s
	clock.AdvanceBy(10 * time.Second)

	// assert: the ticker should tick 10 times in 10s
	Expect(cnt.Load(), "ticks").To(Equal[uint32](10))
}

func TestMock_DropsTicks(t *testing.T) {
	With(t)

	// arrange: establish a mock clock with DropsTicks set, create
	// a ticker to tick every 1s and start a goroutine to count the
	// ticks from the ticker
	var (
		cnt   atomic.Uint32
		clock = internal.NewMockClock(internal.DropsTicks())
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	ticker := clock.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				cnt.Add(1)
			case <-done:
				return
			}
		}
	}()

	// act: advance the clock by 10s
	clock.AdvanceBy(10 * time.Second)

	// assert: the ticker would ordinarily tick 10 times in 10s, but
	// with DropsTicks it should only tick once
	Expect(cnt.Load(), "ticks").To(Equal[uint32](1))
}

func TestMock_panicIfLocked_WhenLocked(t *testing.T) {
	With(t)

	// arrange: create a mock clock and lock it
	clock := internal.NewMockClock()
	clock.Lock()
	defer Expect(Panic(internal.ErrClockLocked)).DidOccur()

	// act/assert: creating a ticker attempt to lock the clock again (should panic)
	clock.NewTicker(1 * time.Second)
}

// Tests that a mocked ContextWithDeadline is cancelled when the mock clock is
// advanced to the deadline.
func Test_Mocked_ContextWithDeadline(t *testing.T) {
	With(t)

	var (
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	ctx, _ = clock.ContextWithDeadline(ctx, clock.Now().Add(time.Second))

	clock.AdvanceBy(time.Second)
	select {
	case <-ctx.Done():
		Expect(ctx.Err()).Is(context.DeadlineExceeded)
	default:
		Fatal("context is not cancelled when deadline exceeded")
	}
}

// Tests that the mocked ContextWithDeadlineCause Stringer describes
// the context correctly.
func Test_Mocked_ContextWithDeadlineCause_Stringer(t *testing.T) {
	With(t)

	var (
		clock       = internal.NewMockClock()
		deadline    = clock.Now().Add(time.Second)
		cause       = errors.New("cause")
		ctx, _      = clock.ContextWithDeadlineCause(context.Background(), deadline, cause)
		expectedErr = fmt.Sprintf("%s: %s", context.DeadlineExceeded, cause)

		asString = func(ctx context.Context) string {
			return fmt.Sprintf("%s", ctx)
		}
	)

	const (
		deadlineFormat         = "MockContext{deadline: %s (in 1s), cause: %q}"
		deadlineExceededFormat = "MockContext{deadline: %s (in 0s), err: %q}"
	)

	Expect(asString(ctx)).To(Equal(fmt.Sprintf(deadlineFormat, deadline, cause)))

	clock.AdvanceBy(time.Second)

	Expect(asString(ctx)).To(Equal(fmt.Sprintf(deadlineExceededFormat, deadline, expectedErr)))
}

// Tests that a mocked ContextWithDeadlineCause wraps the cause error
// and is cancelled when the mock clock is advanced to the deadline.
func Test_Mocked_ContextWithDeadlineCause(t *testing.T) {
	With(t)

	// arrange
	var (
		cause = errors.New("cause")
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	// act: create a context with a deadline and a cause then advance the clock
	//      to the deadline
	ctx, _ = clock.ContextWithDeadlineCause(ctx, clock.Now().Add(time.Second), cause)
	clock.AdvanceBy(time.Second)

	// assert: the context should be cancelled with the cause error
	select {
	case <-ctx.Done():
		Expect(ctx.Err()).Is(context.DeadlineExceeded)
		Expect(ctx.Err().Error()).To(Equal(fmt.Sprintf("%s: %s", context.DeadlineExceeded, cause)))
	default:
		Fatal("context was not cancelled")
	}
}

// Tests that a mocked ContextWithDeadline does nothing when the deadline
// is later than a deadline in the parent context.
func Test_Mocked_ContextWithDeadline_LaterThanParent(t *testing.T) {
	With(t)

	// arrange
	var (
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	// act: create a context with a deadline that is later than the parent
	//      then advance the clock to the parent deadline
	ctx, _ = clock.ContextWithDeadline(ctx, clock.Now().Add(time.Second))
	ctx, _ = clock.ContextWithDeadline(ctx, clock.Now().Add(10*time.Second))
	clock.AdvanceBy(time.Second)

	// assert: the context should be cancelled with deadline exceeded
	select {
	case <-ctx.Done():
		Expect(ctx.Err()).Is(context.DeadlineExceeded)
	default:
		Fatal("context was not cancelled")
	}
}

// Tests that the cancel func returned by a mocked ContextWithDeadline cancels
// the context correctly without needing to advance the clock.
func Test_Mocked_ContextWithDeadline_Cancel(t *testing.T) {
	With(t)

	// arrange
	var (
		dur   = 10 * time.Millisecond
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	// act: create a context with a deadline and cancel it immediately
	ctx, cancel := clock.ContextWithDeadline(ctx, clock.Now().Add(dur))
	cancel()

	// assert
	select {
	case <-ctx.Done():
		Expect(ctx.Err()).Is(context.Canceled)
	case <-time.After(dur):
		Fatal("context was not cancelled")
	}
}

// Tests that a mocked ContextWithDeadline cancels a context if the parent context
// is cancelled.
func Test_Mocked_ContextWithDeadline_ParentCancelled(t *testing.T) {
	With(t)

	// arrange
	var (
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	// act: create a cancelable parent context and a child context with a deadline
	//      then cancel the parent context
	parent, cancelParent := context.WithCancel(ctx)
	child, _ := clock.ContextWithDeadline(parent, clock.Now().Add(time.Second))
	cancelParent()

	// assert: the child context should be cancelled
	select {
	case <-child.Done():
		Expect(child.Err()).Is(context.Canceled)
	case <-time.After(time.Second):
		Fatal("child context was not cancelled")
	}
}

// Tests that a mocked ContextWithDeadline does not cancel parent when cancelled before
// the parent.
func Test_Mocked_ContextWithDeadline_ChildCancelled(t *testing.T) {
	With(t)

	// arrange
	var (
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	// act: create a parent context with deadline and a child context with an earlier
	//      dedaline, then cancel the child context
	parent, cancelParent := clock.ContextWithDeadline(ctx, clock.Now().Add(10*time.Millisecond))
	defer cancelParent()

	child, cancelChild := clock.ContextWithDeadline(parent, clock.Now().Add(5*time.Millisecond))
	cancelChild()

	// assert: the child context is cancelled, the parent context is not
	select {
	case <-child.Done():
		Expect(child.Err()).Is(context.Canceled)
		Expect(parent.Err()).Is(nil)
	default:
		Fatal("child context was not cancelled")
	}

	// act: advance the clock to the deadline of the parent
	clock.AdvanceBy(10 * time.Millisecond)

	// assert: the parent context is now expired
	select {
	case <-parent.Done():
		Expect(parent.Err()).Is(context.DeadlineExceeded)
	default:
		Fatal("parent context was not cancelled")
	}
}

// Tests that a mock ContextWithDeadline with a deadline that has already passed
// is cancelled immediately and returns a no-op cancel function.
func Test_Mocked_ContextWithDeadline_DeadlineAlreadyPassed(t *testing.T) {
	With(t)

	// arrange
	var (
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	// act: create a context with a deadline in the past
	ctx, cancel := clock.ContextWithDeadline(ctx, clock.Now().Add(-time.Second))

	// assert: the context is cancelled immediately
	select {
	case <-ctx.Done():
		Expect(ctx.Err()).Is(context.DeadlineExceeded)
	case <-time.After(time.Millisecond):
		Fatal("context was not immediately cancelled")
	}

	// act: cancel the context
	cancel()

	// assert: cancellation did not change the error
	Expect(ctx.Err()).Is(context.DeadlineExceeded)
}

// Tests that a context created using ContextWithTimeout is cancelled when
// deadline is reached.
func Test_Mocked_ContextWithTimeout(t *testing.T) {
	With(t)

	// arrange
	var (
		clock = internal.NewMockClock()
		ctx   = context.Background()
	)

	// act: create a context with a timeout of 1 second then advance the clock
	//	    to expire the timeout
	ctx, _ = clock.ContextWithTimeout(ctx, time.Second)
	clock.AdvanceBy(time.Second)

	// assert: the context should be cancelled with deadline exceeded
	select {
	case <-ctx.Done():
		Expect(ctx.Err()).Is(context.DeadlineExceeded)
	default:
		Fatal("context was not cancelled")
	}
}
