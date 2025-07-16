package time

import (
	"time"

	"github.com/blugnu/time/internal"
)

// Dur returns the duration taken to execute the function f.
// It uses the [SystemClock] to measure the duration.
func Dur(f func()) time.Duration {
	clock := internal.SystemClockInstance
	start := clock.Now()
	f()
	return clock.Since(start)
}
