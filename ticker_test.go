package time

import (
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestTicker_EnterState_InvalidState(t *testing.T) {
	ticker := &ticker{}
	defer test.ExpectPanic(errInvalidState).Assert(t)

	// act: attempt to enter invalid state
	ticker.enterState(99)
}

func TestTicker_EnterState_NoTransition(t *testing.T) {
	ticker := &ticker{state: tsStopped}
	defer test.ExpectPanic(nil).Assert(t)

	// act: attempt to enter active state, which is the default state
	ticker.enterState(tsStopped)
}

func TestTicker_EnterState_Expired(t *testing.T) {
	ticker := &ticker{}
	defer test.ExpectPanic(errInvalidTransition).Assert(t)

	// act: enter expired state
	ticker.enterState(tsExpired)
}

func TestTicker_Reset_NotInitialized(t *testing.T) {
	ticker := &Ticker{}
	defer test.ExpectPanic(errResetCalledOnUninitialized).Assert(t)

	ticker.Reset(time.Second)
}

func TestTicker_Reset_ZeroDuration(t *testing.T) {
	ticker := NewMockClock().NewTicker(1 * time.Millisecond)
	defer test.ExpectPanic(errNonPositiveInterval).Assert(t)

	ticker.Reset(0)
}

func TestTicker_Tick_WhenNil(t *testing.T) {
	var sut *ticker

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	test.IsFalse(t, result)
}

func TestTicker_Tick_WhenNotActive(t *testing.T) {
	var sut = &ticker{state: tsStopped}

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	test.IsFalse(t, result)
}

func TestTicker_Tick_NextInTheFuture(t *testing.T) {
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
	test.IsFalse(t, result)
}
