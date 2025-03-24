package time

import (
	"time"
)

// AtNow is a convenience for AtTime(time.Now()).
//
// This may be useful for testing purposes when you want to start the clock at the
// current time whilst retaining the ability to control the advancement of time.
//
// The time is set in the location of the clock.
func AtNow() ClockOption {
	return AtTime(time.Now())
}

// AtTime sets the initial time of the mock clock.
//
// This may be useful for testing purposes when you want to start the clock at a
// particular time whilst retaining the ability to control the advancement of time.
// The time is set in the location of the clock.
//
// # Default
//
//	1970-01-01 00:00:00 +0000 UTC (in the location of the clock)
func AtTime(t time.Time) ClockOption {
	return func(m *mockClock) {
		m.now = t.In(m.loc)
		m.updated = time.Now()
	}
}

// DropsTicks sets the mock clock to drop ticks when the clock is advanced.
// That is, if the clock is advanced by a duration that would ordinarily
// result in a ticker being triggered more than once, the clock will only
// trigger a single tick event for the final tick.
//
// # Example
//
// When a ticker is set to tick every 300 ms and the clock is advanced by 1s:
//
//   - in normal operation, the ticker will be triggered at 300ms, 600ms
//     and 900ms.
//
//   - with DropsTicks applied, the clock will only send a tick event for
//     the final tick at 900ms.
//
// This may be used to simulate a reader that is reading from a Ticker and
// failing to "keep up".  This is not ideal since it is using the clock to
// simulate the reader behaviour but may be easier than contriving that
// reader behaviour in other ways for testing purposes.
//
// # Default
//
//	not set/disabled
func DropsTicks() ClockOption {
	return func(m *mockClock) {
		m.dropsTicks = true
	}
}

// InLocation sets the mock clock to the given location; the time returned by
// Now() will be in this location.
//
// It is not normally necessary to set the location of the clock but may be
// useful when you want to start the clock in a particular location whilst
// retaining the ability to control the advancement of time.
//
// # Default
//
//	UTC
func InLocation(loc *time.Location) ClockOption {
	return func(m *mockClock) {
		m.now = m.now.In(loc)
	}
}

// StartRunning sets the mock clock to start in a running state.  In this state
// the clock is advanced by elapsed time whenever Now() is obtained from the
// clock or when Update() is explicitly called.
//
// AdvanceBy() and AdvanceTo() are not supported when the clock is in a
// running state and will panic.
//
// This more closely mimics the behaviour of a real clock but means that
// tests will run in real-time; this is not recommended for most tests as
// it will make them run more slowly than they might.
//
// # Default
//
//	not set / stopped
func StartRunning() ClockOption {
	return func(m *mockClock) {
		m.Start()
	}
}

// Yielding sets a duration for which the calling goroutine will be suspended
// when performing operations such as advancing the clock or adding a timer or ticker.
//
// This allows other goroutines to be scheduled at times when it may be useful for a test.
// The duration should rarely need to be changed and should not be set to a value that is
// too high as this will cause a test to run more slowly than it might.
//
// To disable this behaviour (not recommended) set the duration to 0.
//
// # Default
//
//	1ms
func Yielding(d time.Duration) ClockOption {
	return func(m *mockClock) {
		m.yield = max(d, 0)
	}
}
