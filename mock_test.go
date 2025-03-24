package time

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blugnu/test"
)

// Test that a ticker established by After sends at the correct time.
func TestMock_After(t *testing.T) {
	var (
		clock    = NewMockClock()
		ticked   atomic.Bool
		listener WaitFuncs
	)

	// Create a channel to execute after 10 mock seconds.
	ch := clock.After(10 * time.Second)
	listener.Go(func() {
		<-ch
		ticked.Store(true)
	})

	// Move clock forward to just before the time.
	clock.AdvanceBy(9 * time.Second)
	test.IsFalse(t, ticked.Load(), "fired early")

	// Move clock forward to the after channel's time.
	clock.AdvanceBy(1 * time.Second)
	listener.Wait()
	test.IsTrue(t, ticked.Load(), "fired on time")
}

// Ensure that the mock's After channel doesn't block on write.
func TestMock_UnusedAfter(t *testing.T) {
	mock := NewMockClock()
	mock.After(1 * time.Millisecond)

	done := make(chan bool, 1)
	go func() {
		mock.AdvanceBy(1 * time.Second)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("mock.AdvanceBy hung")
	}
}

// Ensure that the mock's AfterFunc executes at the correct time.
func TestMock_AfterFunc(t *testing.T) {
	var ticked atomic.Bool
	clock := NewMockClock()

	// Execute function after duration.
	clock.AfterFunc(10*time.Second, func() {
		ticked.Store(true)
	})

	// Move clock forward to just before the time.
	clock.AdvanceBy(9 * time.Second)
	test.IsFalse(t, ticked.Load(), "fired early")

	// Move clock forward to the after channel's time.
	clock.AdvanceBy(1 * time.Second)
	test.IsTrue(t, ticked.Load(), "fired on time")
}

// Ensure that the mock's AfterFunc doesn't execute if stopped.
func TestMock_AfterFunc_Stop(t *testing.T) {
	// Execute function after duration.
	clock := NewMockClock()
	timer := clock.AfterFunc(10*time.Second, func() {
		t.Fatal("unexpected function execution")
	})

	// Stop timer & move clock forward.
	timer.Stop()
	clock.AdvanceBy(10 * time.Second)
}

// Ensure that the mock's current time can be changed.
func TestMock_Now(t *testing.T) {
	clock := NewMockClock()
	if now := clock.Now(); !now.Equal(time.Unix(0, 0)) {
		t.Fatalf("expected epoch, got: %v", now)
	}

	// Add 10 seconds and check the time.
	clock.AdvanceBy(10 * time.Second)
	if now := clock.Now(); !now.Equal(time.Unix(10, 0)) {
		t.Fatalf("expected epoch, got: %v", now)
	}
}

// Test that IsRunning returns the state of the clock.
func TestMock_IsRunning(t *testing.T) {
	// arrange: create a clock in default (stopped) state
	clock := NewMockClock()

	// assert: that the clock is not running
	test.IsFalse(t, clock.IsRunning())

	// act/assert: start the clock and check that it is running
	clock.Start()
	test.IsTrue(t, clock.IsRunning())
}

// Test that IsRunning returns false when the clock is not running.
func TestMock_IsRunning_StoppedClock(t *testing.T) {
	// arrange: create a clock in stopped state
	clock := NewMockClock()

	// act/assert: check that the clock is not running
	test.IsFalse(t, clock.IsRunning())
}

//

func TestMock_Since(t *testing.T) {
	t.Run("advancing a frozen clock", func(t *testing.T) {
		clock := NewMockClock()

		beginning := clock.Now()
		clock.AdvanceBy(500 * time.Second)

		test.That(t, clock.Since(beginning).Seconds()).Equals(500)
	})

	t.Run("with running clock", func(t *testing.T) {
		clock := NewMockClock(StartRunning())
		time.Sleep(25 * time.Millisecond)

		test.IsTrue(t, clock.Since(time.Unix(0, 0)).Milliseconds() >= 25)
	})
}

func TestMock_Until(t *testing.T) {
	clock := NewMockClock()

	end := clock.Now().Add(500 * time.Second)
	if dur := clock.Until(end); dur.Seconds() != 500 {
		t.Fatalf("expected 500s duration between `clock` and `end`, actually: %v", dur.Seconds())
	}
	clock.AdvanceBy(100 * time.Second)
	if dur := clock.Until(end); dur.Seconds() != 400 {
		t.Fatalf("expected 400s duration between `clock` and `end`, actually: %v", dur.Seconds())
	}
}

// Test that Sleep respects the passage of mocked time for a stopped clock.
func TestMock_Sleep_StoppedClock(t *testing.T) {
	// arrange: start a goroutine that sleeps for 10 seconds
	var (
		clock = NewMockClock()
		ok    atomic.Bool
	)
	go func() {
		clock.Sleep(10 * time.Second)
		ok.Store(true)
	}()

	// act/assert: after 9 mock seconds, the goroutine should still be sleeping
	clock.AdvanceBy(9 * time.Second)
	test.IsFalse(t, ok.Load(), "woke early")

	// act/assert: after 1 more second, the goroutine should have awoken
	clock.AdvanceBy(1 * time.Second)
	test.IsTrue(t, ok.Load(), "woke when expected")
}

// Tests that negative Sleep returns immediately when clock is running.
func TestMock_Sleep_Negative_RunningClock(t *testing.T) {
	// arrange: create a clock in running state and sleep for -1ms
	var (
		clock = NewMockClock(StartRunning())
		start = time.Now()
		dur   time.Duration
	)
	WaitFor(func() {
		clock.Sleep(-1 * time.Millisecond)
		dur = time.Since(start)
	})

	// act/assert: the clock should not have advanced
	test.IsTrue(t, dur < 1*time.Millisecond)
}

// Tests that negative Sleep returns immediately when clock is stopped.
func TestMock_Sleep_Negative_StoppedClock(t *testing.T) {
	// arrange: create a clock in stopped state and sleep for -1ms
	var (
		clock = NewMockClock()
	)
	WaitFor(func() {
		clock.Sleep(-1 * time.Millisecond)
	})

	// act/assert: the clock should not have advanced
	test.IsTrue(t, clock.SinceCreated() == 0)
}

// Tests that Sleep respects the passage of elapsed time for a running clock.
func TestMock_Sleep_RunningClock(t *testing.T) {
	// arrange: create a clock in running state and sleep for 10ms
	clock := NewMockClock(StartRunning())
	time.Sleep(10 * time.Millisecond)

	// act/assert: the clock should have advanced by at least 10ms
	test.IsTrue(t, clock.SinceCreated() >= 10*time.Millisecond)
}

// Tests that a zero Tick duration returns a nil channel.
func TestMock_Tick_Zero(t *testing.T) {
	// arrange: create a clock and a channel to receive ticks
	clock := NewMockClock()
	tick := clock.Tick(0)

	// act/assert: the tick channel should be nil
	test.IsNil(t, tick)
}

// Tests that a negative Tick duration returns a nil channel.
func TestMock_Tick_Negative(t *testing.T) {
	// arrange: create a clock and a channel to receive ticks
	clock := NewMockClock()
	tick := clock.Tick(-1 * time.Second)

	// act/assert: the tick channel should be nil
	test.IsNil(t, tick)
}

// Tests that Start resumes a stopped clock.
func TestMock_Start(t *testing.T) {
	// arrange: create a default (stopped) clock
	clock := NewMockClock()

	// act/assert: sleep for 10ms and check that the clock has not advanced
	time.Sleep(10 * time.Millisecond)
	test.IsTrue(t, clock.SinceCreated() == 0)

	// act: start the clock and sleep for another 10ms
	clock.Start()
	time.Sleep(10 * time.Millisecond)

	// assert: the mock time should reflect the passing of at least 10ms in real time
	test.IsTrue(t, clock.SinceCreated() >= 10*time.Millisecond)
}

// Tests that Start panics if the clock is already running.
func TestMock_Start_Running(t *testing.T) {
	// arrange/assert: create a default (running) clock
	clock := NewMockClock(StartRunning())
	defer test.ExpectPanic(ErrClockIsRunning).Assert(t)

	// act: attempt to start the clock (again)
	clock.Start()
}

// Tests that Stop pauses a running clock.
func TestMock_Stop(t *testing.T) {
	// arrange: create a default (running) clock
	clock := NewMockClock(StartRunning())

	// assert: verify that the clock is running;
	//  sleep for 10ms and ensure that mock time reflects the elapsed time
	time.Sleep(10 * time.Millisecond)
	test.IsTrue(t, clock.SinceCreated() >= 10*time.Millisecond)

	// act/assert: stop the clock, record the mock time then sleep for 10ms
	//  and verify that the mock time has not advanced
	clock.Stop()
	stoppedAt := clock.Now()
	time.Sleep(10 * time.Millisecond)
	test.IsTrue(t, clock.Since(stoppedAt) == 0)
}

// Tests that a channel established by Tick sends at the correct time.
func TestMock_Tick(t *testing.T) {
	// arrange: start a ticker to fire every 10 seconds and count the ticks
	var (
		clock = NewMockClock()
		ticks atomic.Uint32
	)
	tick := clock.Tick(10 * time.Second)
	go func() {
		for {
			<-tick
			ticks.Add(1)
		}
	}()

	// act/assert: there should be no ticks until the clock is advanced to the
	// first tick time
	clock.AdvanceBy(9 * time.Second)
	test.Value(t, ticks.Load()).Equals(0)

	// act/assert: after 1 more second, the first tick should have fired
	clock.AdvanceBy(1 * time.Second)
	test.Value(t, ticks.Load()).Equals(1)

	// act/assert: after 20 more seconds there should have been 2 further ticks
	clock.AdvanceBy(20 * time.Second)
	test.Value(t, ticks.Load()).Equals(3)
}

// Tests that a Ticker channel sends at the correct time.
func TestMock_Ticker(t *testing.T) {
	var cnt atomic.Uint32
	clock := NewMockClock()

	// Create a channel to increment every microsecond.
	go func() {
		ticker := clock.NewTicker(1 * time.Microsecond)
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(10 * time.Microsecond)
	test.Value(t, cnt.Load()).Equals(10)
}

// Tests that a Ticker with zero duration fires immediately.
func TestMock_Ticker_Zero(t *testing.T) {
	// arrange: create a clock and a channel to receive ticks
	clock := NewMockClock()
	ticker := clock.NewTicker(0)

	// assert: the tick channel should not be nil
	test.IsNotNil(t, ticker.C)

	// assert: the ticker ticked at the clock creation time
	tm := <-ticker.C
	test.IsTrue(t, clock.Since(tm) == 0)
}

// Ensure that the mock's Ticker channel won't block if not read from.
func TestMock_Ticker_Overflow(t *testing.T) {
	clock := NewMockClock()
	ticker := clock.NewTicker(1 * time.Microsecond)
	clock.AdvanceBy(10 * time.Microsecond)
	ticker.Stop()
}

// Ensure that the mock's Ticker can be stopped.
func TestMock_Ticker_Stop(t *testing.T) {
	var cnt atomic.Uint32
	clock := NewMockClock()

	// Create a channel to increment every second.
	ticker := clock.NewTicker(1 * time.Second)
	go func() {
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(5 * time.Second)
	test.Value(t, cnt.Load()).Equals(5)

	ticker.Stop()

	// Move clock forward again.
	clock.AdvanceBy(5 * time.Second)
	test.Value(t, cnt.Load()).Equals(5)
}

func TestMock_Ticker_Reset(t *testing.T) {
	var cnt atomic.Uint32
	clock := NewMockClock()

	ticker := clock.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(10 * time.Second)
	test.Value(t, cnt.Load()).Equals(2)

	clock.AdvanceBy(4 * time.Second)
	ticker.Reset(5 * time.Second)

	// Advance the remaining second
	clock.AdvanceBy(1 * time.Second)

	test.Value(t, cnt.Load()).Equals(2)

	// Advance the remaining 4 seconds from the previous tick
	clock.AdvanceBy(4 * time.Second)

	test.Value(t, cnt.Load()).Equals(3)
}

func TestMock_Ticker_Stop_Reset(t *testing.T) {
	clock := NewMockClock()

	ticker := clock.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var cnt atomic.Uint32
	go func() {
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()

	// Move clock forward.
	clock.AdvanceBy(10 * time.Second)
	test.Value(t, cnt.Load()).Equals(2)

	ticker.Stop()

	// Move clock forward again.
	clock.AdvanceBy(5 * time.Second)
	test.Value(t, cnt.Load()).Equals(2)

	ticker.Reset(2 * time.Second)

	// Advance the remaining 2 seconds
	clock.AdvanceBy(2 * time.Second)

	test.Value(t, cnt.Load()).Equals(3)

	// Advance another 2 seconds
	clock.AdvanceBy(2 * time.Second)

	test.Value(t, cnt.Load()).Equals(4)
}

// Ensure that multiple tickers can be used together.
func TestMock_Ticker_Multi(t *testing.T) {
	var cnt atomic.Uint32
	clock := NewMockClock()

	go func() {
		ones := clock.NewTicker(1 * time.Microsecond)
		tens := clock.NewTicker(3 * time.Microsecond)

		for {
			select {
			case <-ones.C:
				cnt.Add(1)
			case <-tens.C:
				cnt.Add(10)
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
	test.Value(t, cnt.Load()).Equals(40)
}

func TestMock_Timer_Reset_Zero(t *testing.T) {
	// arrange: create a clock and a timer
	var (
		clock    = NewMockClock()
		timer    = clock.NewTimer(1 * time.Second)
		ticked   atomic.Bool
		dur      time.Duration
		listener WaitFuncs
	)
	listener.Go(func() {
		dur = clock.Since(<-timer.C)
		ticked.Store(true)
	})
	timer.Reset(0)
	listener.Wait()

	// act/assert: the timer should have fired immediately
	test.Value(t, dur).Equals(0)
}

func TestMock_ReentrantDeadlock(t *testing.T) {
	mockedClock := NewMockClock()
	timer20 := mockedClock.NewTimer(20 * time.Second)
	go func() {
		v := <-timer20.C
		panic(fmt.Sprintf("timer should not have ticked: %v", v))
	}()
	mockedClock.AfterFunc(10*time.Second, func() {
		timer20.Stop()
	})

	mockedClock.AdvanceBy(15 * time.Second)
	mockedClock.AdvanceBy(15 * time.Second)
}

// Test that a running clock advances by the elapsed time
func TestMock_Advance(t *testing.T) {
	// arrange: create a mock clock in running state
	clock := NewMockClock(StartRunning())

	// act: sleep for 100ms then advance the clock
	clock.Sleep(100 * time.Millisecond)
	clock.Update()

	// assert: the clock should have advanced by at least 100ms
	test.IsTrue(t, clock.Since(time.Time{}) >= 100*time.Millisecond)
}

// Tests that Advance panics if the clock is not running.
func TestMock_Advance_NotRunning(t *testing.T) {
	// arrange: create a mock clock in stopped state
	clock := NewMockClock()

	// act/assert: attempt to advance the clock (should panic)
	defer test.ExpectPanic(ErrClockNotRunning).Assert(t)
	clock.Update()
}

// Tests that AdvanceBy advances the clock by the specified duration.
func TestMock_AdvanceBy(t *testing.T) {
	// arrange: create a mock clock
	clock := NewMockClock()

	// act: advance the clock by 100ms
	clock.AdvanceBy(100 * time.Millisecond)

	// assert: the clock should have advanced by at least 100ms
	test.IsTrue(t, clock.SinceCreated() == 100*time.Millisecond)
}

// Tests that AdvanceBy panics if attempting to go back in time.
func TestMock_AdvanceBy_GoingBackInTime(t *testing.T) {
	// arrange: create a mock clock
	clock := NewMockClock()

	// act/assert: attempt to advance the clock back in time
	defer test.ExpectPanic(ErrNotADelorean).Assert(t)
	clock.AdvanceBy(-100 * time.Millisecond)
}

// Tests that AdvanceTo advances the clock to the specified time.
func TestMock_AdvanceTo(t *testing.T) {
	// arrange: create a mock clock
	clock := NewMockClock()

	// act: advance the clock to a specific time
	clock.AdvanceTo(time.Unix(100, 0))

	// assert: the clock should have advanced to the specified time
	test.IsTrue(t, clock.Now().Equal(time.Unix(100, 0)))
}

// Tests that AdvanceTo panics if attempting to go back in time.
func TestMock_AdvanceTo_GoingBackInTime(t *testing.T) {
	// arrange: create a mock clock
	clock := NewMockClock()

	// act/assert: attempt to advance the clock back in time
	defer test.ExpectPanic(ErrNotADelorean).Assert(t)
	clock.AdvanceTo(time.Unix(-100, 0))
}

// Test that many simultaneous timers can be created and that they
// all tick at the correct time.
func TestMock_AfterFuncRace(t *testing.T) {
	var (
		clock  = NewMockClock()
		called atomic.Bool
	)
	defer func() {
		test.IsTrue(t, called.Load(), "func is called")
	}()

	funcs := StartFuncs{}
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
	// arrange: prepare a number of goroutines to setup timers to tick after 1ms.
	// The goroutines will all be started at the same time and will all set their timers
	// from the same base time.

	const n = 20
	var (
		mock  = NewMockClock()
		ticks atomic.Int32
		funcs StartFuncs
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
	test.Value(t, ticks.Load(), "ticks").Equals(n)
}

func TestMock_DoesNotDropTicks(t *testing.T) {
	// arrange: establish a mock clock with DropsTicks set, create
	// a ticker to tick every 1s and start a goroutine to count the
	// ticks from the ticker
	clock := NewMockClock()
	var cnt atomic.Uint32

	ticker := clock.NewTicker(1 * time.Second)

	go func() {
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()

	// act: advance the clock by 10s
	clock.AdvanceBy(10 * time.Second)

	// assert: the ticker should tick 10 times in 10s
	test.Value(t, cnt.Load(), "ticks").Equals(10)
}

func TestMock_DropsTicks(t *testing.T) {
	// arrange: establish a mock clock with DropsTicks set, create
	// a ticker to tick every 1s and start a goroutine to count the
	// ticks from the ticker
	clock := NewMockClock(DropsTicks())
	var cnt atomic.Uint32

	ticker := clock.NewTicker(1 * time.Second)

	go func() {
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()

	// act: advance the clock by 10s
	clock.AdvanceBy(10 * time.Second)

	// assert: the ticker would ordinarily tick 10 times in 10s, but
	// with DropsTicks it should only tick once
	test.Value(t, cnt.Load(), "ticks").Equals(1)
}

func TestMock_panicIfLocked_WhenLocked(t *testing.T) {
	// arrange: create a mock clock and lock it
	clock := NewMockClock().(*mockClock)
	clock.Lock()
	defer test.ExpectPanic(errClockLocked).Assert(t)

	// act/assert: attempt to lock the clock again (should panic)
	clock.panicIfLocked()
}
