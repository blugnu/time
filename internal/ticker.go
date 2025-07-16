package internal

import (
	"fmt"
	"time"
)

// Ticker implements a ticker that can be used with a mock clock
type Ticker struct {
	// wraps a time.Timer in normal use; for a mock, this is non-nil but is
	// used only as a container for the <-chan time.Time read-only reference
	// to the mock timer's channel.
	*time.Ticker

	// non-nil only when timer is mocked
	*ticker

	// indicates whether the ticker has been initialized
	initialised bool
}

// Reset resets the ticker to the specified duration.
//
// If the Ticker has been stopped it is restarted with the new duration.
//
// If the Ticker is already running it will be reset to the new duration; the
// next tick will occur at the specified duration from the current time.
//
// The function panics if the given duration is zero or negative, or if the
// Ticker has not been initialized.
func (t *Ticker) Reset(d time.Duration) {
	if !t.initialised {
		panic(fmt.Errorf("%w Ticker", errResetCalledOnUninitialized))
	}

	if t.isMocked() {
		t.ticker.reset(d)
		return
	}

	t.Ticker.Reset(d)
}

// Stop stops the ticker and prevents any further ticks from being sent to
// the channel; the channel is not closed.
func (t *Ticker) Stop() {
	switch {
	case t.isMocked():
		t.ticker.stop()
	case t.initialised:
		t.Ticker.Stop()
	}
}

// ticker implements the behaviour of a Ticker using a mock clock.
type ticker struct {
	tickerId int
	c        chan time.Time
	d        time.Duration
	next     time.Time
	state    tickerState
	clock    *MockClock
}

// id returns the id of the ticker.
func (mock *ticker) id() int {
	return mock.tickerId
}

// enterState handles the transition of the ticker to a new state.
// It will panic if the transition is invalid or if the state is not
// supported by the ticker.
func (mock *ticker) enterState(state tickerState) {
	if mock.state == state {
		return
	}
	mock.state = state

	switch state {
	case tsActive:
		mock.clock.enableTicker(mock.tickerId)
	case tsStopped:
		mock.clock.disableTicker(mock.tickerId)
	case tsExpired:
		panic(fmt.Errorf("%w: %s is not supported by a ticker", errInvalidTransition, state))
	default:
		panic(fmt.Errorf("%w: %s", errInvalidState, state))
	}
}

// isMocked returns true if the ticker is a mock ticker
func (t *Ticker) isMocked() bool {
	return t.ticker != nil
}

// nextTick returns the next tick time for the ticker.
func (mock *ticker) nextTick() time.Time {
	return mock.next
}

// reset resets the ticker to the specified duration.
//
// It will panic if the duration is zero or negative, mimicking the behaviour
// of the time.Ticker type in the standard library.
func (t *ticker) reset(d time.Duration) {
	if d <= 0 {
		panic(fmt.Errorf("%w for Ticker", errNonPositiveInterval))
	}
	t.clock.resetTicker(t, d)
}

// stop stops the ticker and prevents any further ticks from being sent to
func (t *ticker) stop() {
	if t.state == tsActive {
		t.enterState(tsStopped)
	}
}

// tick is called to tick the ticker at the given time
// it returns true if the ticker should tick, false otherwise.
func (t *ticker) tick(now time.Time) bool {
	if t == nil || t.state != tsActive || t.next.After(now) {
		return false
	}

	// record the next time at which the tick should occur and update
	// the next tick time to be the next interval
	at := t.next
	t.next = t.next.Add(t.d)

	// if the clock is dropping ticks then we skip forward to the final
	// tick that occurs at/before now
	if t.clock.dropsTicks {
		for !t.next.After(now) {
			at = t.next
			t.next = t.next.Add(t.d)
		}
	}

	// tick at the time that was determined
	go func() {
		clock := t.clock
		clock.Lock()
		clock.now = at
		clock.Unlock()

		t.c <- at
	}()

	// yield to allow goroutines waiting on the ticker channel to be scheduled
	time.Sleep(t.clock.yield)

	return true
}
