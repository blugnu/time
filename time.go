package time

import (
	"context"
	"time"

	"github.com/blugnu/time/internal"
)

// This file provides implementations of package-level symbols provided by the time package.
// In most cases these are aliases for corresponding symbols in the standard time package.

type (
	// the following types are aliases for the corresponding types in the standard time package

	Duration   = time.Duration
	Location   = time.Location
	ParseError = time.ParseError
	Time       = time.Time

	// A Month specifies a month of the year, from January (1) to December (12)
	Month = time.Month

	// A Weekday specifies a day of the week, from Sunday (0) to Saturday (6)
	Weekday = time.Weekday

	// the implementation of these types is replaced by this package to ensure correct
	// behaviour when used with a mock clock

	// Ticker represents a ticker.
	//
	// Usage is identical to the [time.Ticker] type in the standard library: the
	// time of each "tick" is read from the channel `C` provided on the [Ticker].
	//
	// A ticker created using a [MockClock] will tick each time the clock is advanced
	// to (or beyond) the next tick time.
	//
	// If a [MockClock] is advanced by a duration that is greater than the period of
	// the ticker, ticks will be sent at each interval unless the clock was configured
	// with [DropsTicks]. In that case, only the final tick will be produced by the
	// [Ticker].  This may be useful for simulating a tick consumer that is failing
	// to "keep up".
	Ticker = internal.Ticker

	// Timer represents a timer; it may obtained from the SystemClock() or a mock
	// obtained from a MockClock.
	//
	// Usage is identical to the [time.Timer] type in the standard library: the
	// time of the [Timer] is read from the channel `C` provided which is sent
	// when the timer expires or is reset.
	//
	// A [Timer] created from a [MockClock] will tick when the clock is advanced
	// to (or beyond) the time specified on the [Timer].
	Timer = internal.Timer
)

const (
	// durations
	Day         = time.Hour * 24
	Hour        = time.Hour
	Microsecond = time.Microsecond
	Millisecond = time.Millisecond
	Minute      = time.Minute
	Nanosecond  = time.Nanosecond
	Second      = time.Second
	Week        = time.Hour * 24 * 7

	// date/time formats
	Layout = time.Layout // The reference time, in numerical order

	ANSIC       = time.ANSIC
	DateOnly    = time.DateOnly
	DateTime    = time.DateTime
	RFC822      = time.RFC822
	RFC822Z     = time.RFC822Z // RFC822 with numeric zone
	RFC850      = time.RFC850
	RFC1123     = time.RFC1123
	RFC1123Z    = time.RFC1123Z // RFC1123 with numeric zone
	RFC3339     = time.RFC3339
	RFC3339Nano = time.RFC3339Nano
	Kitchen     = time.Kitchen
	RubyDate    = time.RubyDate
	Stamp       = time.Stamp
	StampMilli  = time.StampMilli
	StampMicro  = time.StampMicro
	StampNano   = time.StampNano
	TimeOnly    = time.TimeOnly
	UnixDate    = time.UnixDate
)

var (
	// the following functions are aliases for the corresponding functions in the standard time package

	Date            = time.Date
	FixedZone       = time.FixedZone
	Parse           = time.Parse
	ParseDuration   = time.ParseDuration
	ParseInLocation = time.ParseInLocation
	Unix            = time.Unix
	UnixMicro       = time.UnixMicro
	UnixMilli       = time.UnixMilli

	// the UTC location

	UTC = time.UTC

	// the following functions have no direct equivalent in this package; equivalent functions
	// are provided as methods of a Clock implementation.  Equivalent package functions are
	// provided, requiring a Context parameter which is used to identify the [Clock] to use.
	//
	// If the context does not contain a [Clock], the [SystemClock] will be used.
	//
	// The functions are:
	//
	// After
	// AfterFunc
	// NewTicker
	// NewTimer
	// Now
	// Since
	// Sleep
	// Tick
	// Until
)

// After waits for the duration to elapse and then sends the current time on the returned
// channel. It is equivalent to [NewTimer](d).C.
//
// # Mock Clocks
//
// If the context contains a [MockClock], the channel will not receive a time value until the
// clock is advanced by at least the specified duration from its current time.
//
// Note that a running [MockClock] does not advance automatically; it must be advanced by
// calling [MockClock.Update] or other method which updates the clock after the required
// duration has elapsed in real-time.  i.e. Any goroutines waiting on the channel returned by
// [After] will effectively be blocked in real-time, despite using a [MockClock].
//
// # GODEBUG asynctimerchan=1
//
// The GODEBUG setting asynctimerchan=1 restores pre-Go 1.23 behaviors when using [SystemClock].
// [MockClock] does not reproduce those behaviors and using a [MockClock] with this GODEBUG
// setting is not supported.
//
// For more information, refer to the standard library [time.After] documentation.
func After(ctx context.Context, d Duration) <-chan time.Time {
	return FromContext(ctx).After(d)
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
// It returns a [Timer] that can be used to cancel the call using its Stop method.
// The returned Timer's C field is not used and will be nil.
//
// # Mock Clocks
//
// If the context contains a [MockClock], the function f will not be called until the clock
// is advanced by at least the specified duration from its current time.
//
// Note that a running [MockClock] does not advance automatically; it must be
// advanced by calling [MockClock.Update] or other method which updates the clock
// after the required duration has elapsed in real-time.  i.e. Any goroutines waiting
// on a [Timer] with a running [MockClock] will effectively be blocked in real-time,
// despite using a [MockClock].
func AfterFunc(ctx context.Context, d Duration, f func()) *Timer {
	return FromContext(ctx).AfterFunc(d, f)
}

// NewTicker creates a new Ticker that will send the current time on its channel
// after each tick. The duration d must be greater than zero; if d <= 0, NewTicker
// will panic.
//
// The duration of the Ticker can be modified using the Reset method.
//
// The Ticker will continue ticking until Stop is called on it.
//
// # Mock Clocks
//
// If the context contains a [MockClock], the ticker tick each time the clock is
// advanced by an increment of the specified duration from its current time.
//
// Note that a running [MockClock] does not advance automatically; it must be
// advanced by calling [MockClock.Update] or other method which updates the clock
// after the required duration has elapsed in real-time.  i.e. Any goroutines waiting
// on a [Ticker] with a running [MockClock] will effectively be blocked in real-time,
// despite using a [MockClock].
//
// # GODEBUG asynctimerchan=1
//
// The GODEBUG setting asynctimerchan=1 restores pre-Go 1.23 behaviors when using [SystemClock].
// [MockClock] does not reproduce those behaviors and using a [MockClock] with this GODEBUG
// setting is not supported.
//
// For more information, refer to the standard library [time.NewTicker] documentation.
func NewTicker(ctx context.Context, d Duration) *Ticker {
	return FromContext(ctx).NewTicker(d)
}

// NewTimer creates a new Timer that will send the current time on its channel after
// at least duration d.
//
// # Mock Clocks
//
// If the context contains a [MockClock], the timer will not be triggered until the
// clock is advanced by at least the specified duration from its current time.
//
// Note that a running [MockClock] does not advance automatically; it must be advanced by
// calling [MockClock.Update] or other method which updates the clock after at least the
// required duration has elapsed in real-time to trigger the timer.  i.e. goroutines
// waiting on a [Timer] will effectively be blocked in real-time, despite using a [MockClock].
//
// # GODEBUG asynctimerchan=1
//
// The GODEBUG setting asynctimerchan=1 restores pre-Go 1.23 behaviors when using [SystemClock].
// [MockClock] does not reproduce those behaviors and using a [MockClock] with this GODEBUG
// setting is not supported.
//
// For more information, refer to the standard library [time.NewTimer] documentation.
func NewTimer(ctx context.Context, d Duration) *Timer {
	return FromContext(ctx).NewTimer(d)
}

// Now returns the current time from the [Clock] in the given context. If there is no clock in the
// context, the [SystemClock] will be used.
//
// # Mock Clocks
//
// If the context contains a [MockClock], the time returned will be the current mocked time.
//
// If [MockClock] is running, the time will have advanced by the elapsed time since the clock was
// last updated, e.g. by a call to [Now] or [MockClock.Update].
//
// If the [MockClock] is not running, [Now] will return the same [Time] until the clock is advanced
// by a call to [MockClock.AdvanceBy] or [MockClock.AdvanceTo].
func Now(ctx context.Context) Time {
	return FromContext(ctx).Now()
}

// Tick returns a channel that will send the current time after each tick. If the duration d is
// zero or negative, [Tick] will return a nil channel and will not panic.
//
// Efficiency concerns relating to the [Tick] function prior to Go 1.23 are not significant since
// the blugnu/time package requires at least Go 1.23.
func Tick(ctx context.Context, d Duration) <-chan Time {
	return FromContext(ctx).Tick(d)
}

// Sleep suspends the calling goroutine for the duration specified.
//
// # Mock Clocks
//
// If the context holds a [MockClock], the calling goroutine is suspended until the clock has
// been advanced by at least the specified duration.  The [MockClock] must be advanced by
// a goroutine other than the one calling [Sleep].
//
// Note that a running [MockClock] does not advance automatically; it must be advanced by
// calling [MockClock.Update] or other method which updates the clock after the required duration
// has elapsed in real-time.  i.e. Any goroutines waiting on a [Sleep] call will effectively be
// blocked in real-time, despite using a [MockClock].
func Sleep(ctx context.Context, d Duration) {
	FromContext(ctx).Sleep(d)
}
