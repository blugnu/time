package internal

import (
	"fmt"
	"time"
)

// Timer implements a timer that can be used with a mock clock
type Timer struct {
	// wraps a time.Timer in normal use; for a mock, this is non-nil but is
	// used only as a container for the <-chan time.Time read-only reference
	// to the mock timer's channel.
	*time.Timer

	// non-nil only when timer is mocked
	*timer

	// indicates whether the timer has been initialised
	initialised bool
}

// Reset modifies the timer to expire after duration d from the current time.
// If the timer has already expired it is re-activated.
//
// Returns true if the timer was already active, false if the timer had
// expired or been stopped (and was re-activated).
func (t *Timer) Reset(d time.Duration) bool {
	if !t.initialised {
		panic(fmt.Errorf("%w Timer", errResetCalledOnUninitialized))
	}

	if t.isMocked() {
		return t.timer.reset(d)
	}

	return t.Timer.Reset(d)
}

// Stop prevents the Timer from firing. It returns true if the call stops the
// timer, false if the timer has already expired or been stopped.
func (t *Timer) Stop() bool {
	if t.isMocked() {
		return t.timer.stop()
	}
	return t.Timer.Stop()
}

// timer implements the behaviour of a Timer with a mock clock.
type timer struct {
	tickerId int
	c        chan time.Time
	fn       func()
	next     time.Time
	state    tickerState
	clock    *MockClock
}

// id returns the id of the timer.
func (mock *timer) id() int {
	return mock.tickerId
}

// enterState handles the state transition of the timer.
//
// It will panic if the transition is invalid or if the state is not
// recognized.
func (mock *timer) enterState(state tickerState) {
	if mock.state == state {
		return
	}

	mock.state = state

	switch state {
	case tsActive:
		mock.clock.enableTicker(mock.tickerId)
	case tsExpired:
		mock.clock.disableTicker(mock.tickerId)
	case tsStopped:
		mock.clock.disableTicker(mock.tickerId)
	default:
		panic(fmt.Errorf("%w: %s", errInvalidState, state))
	}
}

// isMocked returns true if the timer is a mock timer, false if it is a
// standard library timer.
func (t *Timer) isMocked() bool {
	return t.timer != nil
}

// nextTick returns the next tick time for the timer.
func (mock *timer) nextTick() time.Time {
	return mock.next
}

// reset modifies the timer to expire after duration d from the current time.
// If the timer has already expired it is re-activated.
// Returns true if the timer was already active, false if the timer had
// expired or been stopped (and was re-activated).
func (t *timer) reset(d time.Duration) bool {
	wasWaiting := t.state == tsActive

	t.clock.resetTimer(t, d)

	return wasWaiting
}

// stop prevents the Timer from firing. It returns true if the call stops the
// timer, false if the timer has already expired or been stopped.
func (t *timer) stop() bool {
	wasActive := t.state == tsActive
	if wasActive {
		t.enterState(tsStopped)
	}

	return wasActive
}

// tick is called to tick the timer at the given time.
func (t *timer) tick(now time.Time) bool {
	if t == nil || t.state != tsActive || t.next.After(now) {
		return false
	}
	t.enterState(tsExpired)

	switch {
	case t.fn != nil:
		go func() { t.clock.now = t.next; t.fn() }()
	case t.c != nil:
		go func() { t.clock.now = t.next; t.c <- t.next }()
	}
	time.Sleep(t.clock.yield)

	return true
}
