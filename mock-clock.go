package time

import (
	"time"

	"github.com/blugnu/time/internal"
)

type MockClockOption = internal.MockClockOption

// MockClock extends the Clock interface with methods to manipulate the
// current time of the clock.  In normal use, the underlying clock time
// is advanced only when explicitly directed to do so using AdvanceBy()
// or AdvanceTo() methods; this is "stopped" mode.
//
// When the clock is "running" the current time is advanced semi-automatically
// by the passage of real-time since the last time the clock was updated.
// In "running" mode, the clock is advanced any time that Now() is called,
// or by calling Update().
//
// It is used to simulate the passage of time in tests.
type MockClock interface {
	// MockClock is a mock implementation of the time.Clock interface.
	Clock

	// AdvanceBy moves the current time of the mock clock forward by a
	// specified duration, triggering any timers or tickers that would have
	// been triggered during that passage of time.
	//
	// Calling this method while the clock is running will result in a panic.
	AdvanceBy(d time.Duration)

	// AdvanceTo moves the current time of the mock clock to a specific time,
	// triggering any timers or tickers that would have been triggered during
	// that passage of time.
	//
	// Calling this method while the clock is running will result in a panic.
	AdvanceTo(t time.Time)

	// CreatedAt returns the mocked time at which the clock was started when created.
	CreatedAt() time.Time

	// IsRunning returns true if the clock is in a running state.
	// In this state the clock is advanced by elapsed time whenever Now()
	// is obtained from the clock or when Update() is explicitly called.
	// AdvanceBy() and AdvanceTo() are not supported when the clock is in a
	// running state and will panic.
	//
	// This more closely mimics the behaviour of a real clock but tests using
	// a running clock may be less deterministic and run more slowly than
	// they might.
	IsRunning() bool

	// SinceCreated returns the elapsed mock time since the clock was created.
	// This is the same as calling clock.Since(clock.CreatedAt()).
	SinceCreated() time.Duration

	// Stop stops the clock from advancing automatically.  Every call to
	// Stop() must be matched with a call to Start() to resume automatic
	// advancement.
	//
	// A MockClock is initially created in stopped mode unless the StartRunning
	// option is specified when initialising the clock.
	Stop()

	// Start resumes automatic advancement of the clock.  Every call to
	// Start() must be matched with a call to Stop() to stop automatic
	// advancement.
	//
	// A MockClock is initially created in stopped mode unless the StartRunning
	// option is specified when initialising the clock.  i.e. if the clock
	// is created in stopped mode, an initial call to Start() is required to
	// start the clock.
	Start()

	// Update moves the current time of the mock clock forward by a duration
	// corresponding to the passage of real-time since it was last updated,
	// triggering any timers or tickers that would have been triggered during
	// that passage of time.
	//
	// Calling this method while the clock is stopped will result in a panic.
	Update()
}

// NewMockClock returns an instance of a mock clock.
//
// The default settings on a new clock are:
//
//   - initial time set to the UNIX epoch (00:00:00 UTC on Thursday, 1 Jan 1970)
//   - stopped; advance with AdvanceBy() or AdvanceTo()
//   - does not drop ticks
//   - sleeps the calling goroutine for 1ms on various operations
//
// When stopped, the clock must be explicitly advanced using AdvanceBy() or
// AdvanceTo().  When not stopped Update() may be used to advance the clock
// by the elapsed real-time since the last advancement.
//
// The clock can be customised using the provided options:
//
//   - AtNow() sets the initial time of the mock clock to the current time;
//
//   - AtTime(t time.Time) sets the initial time of the mock clock;
//
//   - DropsTicks() sets the clock to fire tickers only once where multiple
//     ticks would have been triggered by a single advance of the clock
//
//   - WithYield(d time.Duration) sets a duration for which the calling goroutine is
//     suspended before and after each advancement of the clock.
//
//   - StartRunning() sets the mock clock to start in a running state; in the running
//     state the clock is advanced by elapsed time whenever Now() is obtained from
//     the clock or when Update() is explicitly called.  AdvanceBy() and AdvanceTo()
//     are not supported in the running state and will panic.
func NewMockClock(opts ...internal.MockClockOption) MockClock {
	return internal.NewMockClock(opts...)
}

// AtNow is a convenience for AtTime(time.Now()).
//
// This may be useful for testing purposes when you want to start the clock at the
// current time whilst retaining the ability to control the advancement of time.
//
// The time is set in the location of the clock.
func AtNow() MockClockOption {
	return internal.AtTime(internal.SystemClockInstance.Now())
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
func AtTime(t time.Time) MockClockOption {
	return internal.AtTime(t)
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
func DropsTicks() MockClockOption {
	return internal.DropsTicks()
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
func InLocation(loc *time.Location) MockClockOption {
	return internal.InLocation(loc)
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
func StartRunning() MockClockOption {
	return internal.StartRunning()
}

// YieldTime sets a duration for which the calling goroutine will be suspended
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
func YieldTime(d time.Duration) MockClockOption {
	return internal.YieldTime(d)
}
