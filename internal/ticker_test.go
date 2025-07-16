package internal //nolint:testpackage // tests rely on non-exported symbols

import (
	"sync/atomic"
	"testing"
	"time"

	. "github.com/blugnu/test"
)

func TestTicker(t *testing.T) {
	With(t)

	Run(FlakyTest("ticks at the expected intervals", func() {
		var (
			cnt    atomic.Uint32
			clock  = SystemClockInstance
			ticker = clock.NewTicker(10 * time.Millisecond)
			done   = make(chan struct{})
		)
		defer close(done) // terminates the goroutine we are about to start

		// act: start a goroutine that will count the ticks
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

		// 10ms ticker should tick twice in 25ms
		clock.Sleep(25 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](2))

		// clear the counter and reset the ticker to 20ms
		cnt.Store(0)
		ticker.Reset(20 * time.Millisecond)

		// 20ms ticker should tick once in 25ms
		clock.Sleep(25 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](1))

		// clear the counter and stop the ticker
		cnt.Store(0)
		ticker.Stop()

		// wait another 25ms; the ticker should no longer be ticking
		clock.Sleep(25 * time.Millisecond)
		Expect(cnt.Load(), "ticks").To(Equal[uint32](0))
	}))
}

func TestTicker_EnterState_InvalidState(t *testing.T) {
	With(t)

	ticker := &ticker{}
	defer Expect(Panic(errInvalidState)).DidOccur()

	// act: attempt to enter invalid state
	ticker.enterState(99)
}

func TestTicker_EnterState_NoTransition(t *testing.T) {
	With(t)

	ticker := &ticker{state: tsStopped}
	defer Expect(Panic()).DidNotOccur()

	// act: attempt to enter active state, which is the default state
	ticker.enterState(tsStopped)
}

func TestTicker_EnterState_Expired(t *testing.T) {
	With(t)

	ticker := &ticker{}
	defer Expect(Panic(errInvalidTransition)).DidOccur()

	// act: enter expired state
	ticker.enterState(tsExpired)
}

func TestTicker_Reset_NotInitialized(t *testing.T) {
	With(t)

	ticker := &Ticker{}
	defer Expect(Panic(errResetCalledOnUninitialized)).DidOccur()

	ticker.Reset(time.Second)
}

func TestTicker_Reset_ZeroDuration(t *testing.T) {
	With(t)

	ticker := NewMockClock().NewTicker(1 * time.Millisecond)
	defer Expect(Panic(errNonPositiveInterval)).DidOccur()

	ticker.Reset(0)
}

func TestTicker_Tick_WhenNil(t *testing.T) {
	With(t)

	var (
		sut *ticker
		tm  = time.Now()
	)

	// act: attempt to tick a nil ticker
	result := sut.tick(tm)

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTicker_Tick_WhenNotActive(t *testing.T) {
	With(t)

	var (
		sut = &ticker{state: tsStopped}
		tm  = time.Now()
	)

	// act: attempt to tick a stopped ticker
	result := sut.tick(tm)

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTicker_Tick_NextInTheFuture(t *testing.T) {
	With(t)

	// arrange: setup a ticker with a future tick time
	var (
		clock = NewMockClock()
		sut   = clock.NewTicker(1 * time.Second)
		done  = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	go func() {
		select {
		case <-sut.C:
			Error("ticker should not have ticked")
		case <-done:
			return
		}
	}()

	// act: attempt to tick the timer without advancing the clock
	result := sut.tick(clock.Now())

	// assert: expect false
	Expect(result).To(BeFalse())
}
