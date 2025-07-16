package internal

import "errors"

var (
	// ErrClockIsRunning is returned when Start is called on a mock clock that is already running
	ErrClockIsRunning = errors.New("clock is running")

	// ErrClockNotRunning is returned when Stop is called on a mock clock that is not running.
	// It is also recovered from a panic when a mock clock method is called that requires the clock
	// to be running but is not.
	ErrClockNotRunning = errors.New("clock is stopped")

	// ErrNotADelorean is recovered from a panic resulting from an attempt to send a mock clock
	// back in time.
	ErrNotADelorean = errors.New("not a DeLorean clock (cannot go back in time)")

	// ErrClockLocked is returned when an operation is attempted on a mock clock that is locked
	ErrClockLocked = errors.New("clock is locked")

	errInvalidState      = errors.New("not a valid state")
	errInvalidTransition = errors.New("invalid state transition")

	errNonPositiveInterval        = errors.New("time: non-positive interval")
	errResetCalledOnUninitialized = errors.New("time: Reset called on uninitialized")
)
