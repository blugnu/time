package internal //nolint:testpackage // tests rely on non-exported symbols

import (
	"sync/atomic"
	"testing"
	"time"

	. "github.com/blugnu/test"
)

func TestTimer(t *testing.T) {
	With(t)

	Run(FlakyTest("ticks at the expected intervals", func() {
		var (
			cnt   atomic.Uint32
			clock = SystemClockInstance
			timer = clock.NewTimer(10 * time.Millisecond)
			done  = make(chan struct{})
		)
		defer close(done) // terminates the goroutine we are about to start

		// act: start a goroutine that will count the ticks
		go func() {
			for {
				select {
				case <-timer.C:
					cnt.Add(1)
				case <-done:
					return
				}
			}
		}()

		// 10ms timer should not have ticked after just 6ms...
		clock.Sleep(6 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](0))

		// ...but should have ticked after 12ms
		clock.Sleep(6 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](1))

		// reset the timer to 5ms and clear the counter
		active := timer.Reset(5 * time.Millisecond)
		Expect(active, "timer active").To(BeFalse())

		cnt.Store(0)

		// the new timer should have ticked after 3ms...
		clock.Sleep(3 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](0))

		// ...but should have ticked after 6ms
		clock.Sleep(3 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](1))

		// reset the timer to 10ms and clear the counter
		active = timer.Reset(10 * time.Millisecond)
		Expect(active, "timer active").To(BeFalse())

		cnt.Store(0)

		// stop the timer
		stopped := timer.Stop()
		Expect(stopped, "timer stopped").To(BeTrue())

		// after 15ms the 10ms timer should not have ticked
		// (it was stopped)
		clock.Sleep(15 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](0))
	}))
}

func TestTimer_EnterState_InvalidState(t *testing.T) {
	With(t)

	ticker := &timer{}
	defer Expect(Panic(errInvalidState)).DidOccur()

	ticker.enterState(99)
}

func TestTimer_EnterState_NoTransition(t *testing.T) {
	With(t)

	ticker := &timer{state: tsExpired}
	defer Expect(Panic()).DidNotOccur()

	// act
	ticker.enterState(tsExpired)
}

func TestTimer_Reset_NotInitialized(t *testing.T) {
	With(t)

	timer := &Timer{}
	defer Expect(Panic(errResetCalledOnUninitialized)).DidOccur()

	timer.Reset(time.Second)
}

func TestTimer_Tick_WhenNil(t *testing.T) {
	With(t)

	var sut *timer

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTimer_Tick_WhenNotActive(t *testing.T) {
	With(t)

	var sut = &timer{state: tsExpired}

	// act: attempt to tick an expired timer
	result := sut.tick(time.Time{})

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTimer_Tick_Premature(t *testing.T) {
	With(t)

	var sut = &timer{
		next: time.Now().Add(+1 * time.Second),
	}

	// act
	result := sut.tick(time.Now())

	// assert: expect false
	Expect(result).To(BeFalse())
}
