package time

import (
	"github.com/blugnu/time/internal"
)

// Clock represents an interface described by the functions in the time package
// of the standard library.  It extends the time package with additional
// methods to create contexts with deadlines and timeouts based on the clock
// providing the interface.
//
// This allows for the creation of mock clocks for testing purposes through an
// API that is similar to and consistent with that of the system clock in the
// standard library `time` package.
type Clock = internal.Clock

// SystemClock returns a clock implementation that uses the `time` package functions of the
// standard library. A variable initialised with SystemClock can be used to access the
// system clock functions using function that are identical to those provided by standard
// library time package:
//
//	sys := time.SystemClock()
//	now := sys.Now()
//	sys.Sleep(2 * time.Second)
//	ticker := sys.NewTicker(1 * time.Second)
//
// To migrate applications from using the standard time package to this package, replace
// time package references to an appropriate Clock reference:
//
// # Before:
//
//		import "time"
//
//		func CreateUser(ctx context.Context) {
//		    user := &User{
//		        CreatedAt: time.Now(),
//		    }
//		    ...
//	}
//
// # After:
//
//		import "github.com/blugnu/time"
//
//		func CreateUser(ctx context.Context) {
//		    time := time.FromContext(ctx)
//		    user := &User{
//		        CreatedAt: time.Now(),
//		    }
//		    ...
//	}
//
// Note: the `time` package is shadowed by the `time` variable in the above example.
//
//	This minimises the code changes required to migrate a function but may cause
//	problems or confusion.  Using an alternative variable name avoids this but
//	will require more changes to the code when migrating.
func SystemClock() Clock {
	return internal.SystemClockInstance
}
