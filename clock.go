package time

import "time"

// Clock represents an interface described by the functions in the time package
// of the standard library.
type Clock interface {
	After(d time.Duration) <-chan time.Time
	AfterFunc(d time.Duration, f func()) *Timer
	NewTicker(d time.Duration) *Ticker
	NewTimer(d time.Duration) *Timer
	Now() time.Time
	Since(t time.Time) time.Duration
	Sleep(d time.Duration)
	Tick(d time.Duration) <-chan time.Time
	Until(t time.Time) time.Duration
}

// sysClock is a Clock implementation that wraps the system clock.
var sysClock Clock = systemClock{}

// SystemClock returns a clock implementation that uses the `time` package functions of the
// standard library.
func SystemClock() Clock {
	return sysClock
}

type systemClock struct{}

func (c systemClock) After(d time.Duration) <-chan Time { return time.After(d) }
func (c systemClock) AfterFunc(d time.Duration, f func()) *Timer {
	return &Timer{Timer: time.AfterFunc(d, f)}
}
func (c systemClock) Now() time.Time                        { return time.Now() }
func (c systemClock) Since(t time.Time) time.Duration       { return time.Since(t) }
func (c systemClock) Until(t time.Time) time.Duration       { return time.Until(t) }
func (c systemClock) Sleep(d time.Duration)                 { time.Sleep(d) }
func (c systemClock) Tick(d time.Duration) <-chan time.Time { return time.Tick(d) }

func (c systemClock) NewTicker(d time.Duration) *Ticker {
	return &Ticker{Ticker: time.NewTicker(d), initialised: true}
}

func (c systemClock) NewTimer(d time.Duration) *Timer {
	return &Timer{Timer: time.NewTimer(d), initialised: true}
}
