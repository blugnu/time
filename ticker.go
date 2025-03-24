package time

import (
	"fmt"
	"time"
)

// Ticker represents a ticker; it may be obtained from SystemClock() or a mock
// obtained from a MockClock.
//
// Usage is the same in either case and is identical to the time.Ticker type
// in the standard library: the time of each "tick" is read from the channel `C`
// provided on the Ticker.
//
// A ticker created from a mock clock will tick when the associated mock clock
// is advanced to (or beyond) the next tick time.
//
// If a mock clock is advanced by a duration that is greater than the period of
// the ticker, the ticker will tick at each interval unless the clock was
// configured to drop ticks. In that case, the Ticker will tick only once at
// the last time at/before the time advanced to.
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

func (t *Ticker) isMocked() bool {
	return t.ticker != nil
}

// Reset resets the ticker to the specified duration.
//
// If the Ticker has been stopped it is restarted with the new duration.
//
// If the Ticker is already running it will be reset to the new duration; the
// next tick will occur at the specified duration from the current time.
//
// the function panics if the given duration is zero or negative, or if the
// Ticker has not been initialized, .
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
	if t.isMocked() {
		t.ticker.stop()
		return
	}
	t.Ticker.Stop()
}

// ticker implements the behaviour of a Ticker using a mock clock.
type ticker struct {
	tickerId int
	c        chan time.Time
	d        time.Duration
	next     time.Time
	state    tickerState
	clock    *mockClock
}

// id returns the id of the ticker.
func (mock ticker) id() int {
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

// nextTick returns the next tick time for the ticker.
func (mock ticker) nextTick() time.Time {
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

	// tick at the time that was determined and yield to allow any goroutines
	// that may be waiting on the ticker channel to be scheduled
	go func() { t.clock.withLock(func(c *mockClock) { c.now = at }); t.c <- at }()
	time.Sleep(t.clock.yield)

	return true
}
