package time

import (
	"errors"

	"github.com/blugnu/time/internal"
)

var (
	ErrClockAlreadyExists = errors.New("clock already exists")

	ErrClockIsRunning  = internal.ErrClockIsRunning
	ErrClockNotRunning = internal.ErrClockNotRunning
	ErrNotADelorean    = internal.ErrNotADelorean
)
