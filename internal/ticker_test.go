package internal //nolint:testpackage // tests rely on non-exported symbols

import (
	"sync/atomic"
	"testing"
	"time"

	. "github.com/blugnu/test"
)

func TestTicker(t *testing.T) {
	With(t)

	var cnt atomic.Uint32
	clock := SystemClock{}
	ticker := clock.NewTicker(1 * time.Millisecond)
	go func() {
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()

	clock.Sleep(5300 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](5))

	cnt.Store(0)
	ticker.Reset(5 * time.Millisecond)

	clock.Sleep(10800 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](2))

	cnt.Store(0)
	ticker.Stop()

	clock.Sleep(5900 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](0))
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
	defer Expect(Panic()).DidOccur()

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

	var sut *ticker

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTicker_Tick_WhenNotActive(t *testing.T) {
	With(t)

	var sut = &ticker{state: tsStopped}

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTicker_Tick_NextInTheFuture(t *testing.T) {
	With(t)

	// arrange: setup a ticker with a future tick time
	var (
		clock = NewMockClock()
		sut   = clock.NewTicker(1 * time.Second)
	)
	go func() {
		select {
		case <-sut.C:
			t.Error("should not have ticked")
		default:
			// do nothing
		}
	}()

	// act: attempt to tick the timer without advancing the clock
	result := sut.tick(clock.Now())

	// assert: expect false
	Expect(result).To(BeFalse())
}
