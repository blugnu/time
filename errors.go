package time

import "errors"

var (
	ErrClockAlreadyExists = errors.New("clock already exists")
	ErrClockIsRunning     = errors.New("clock is running")
	ErrClockNotRunning    = errors.New("clock is stopped")
	ErrNotADelorean       = errors.New("not a DeLorean clock (cannot go back in time)")

	errClockLocked       = errors.New("clock is locked")
	errInvalidState      = errors.New("not a valid state")
	errInvalidTransition = errors.New("invalid state transition")

	errNonPositiveInterval        = errors.New("time: non-positive interval")
	errResetCalledOnUninitialized = errors.New("time: Reset called on uninitialized")
)
