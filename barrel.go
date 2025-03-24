package time

import (
	"time"
)

type (
	// the following types are aliases for the corresponding types in the standard time package
	Duration   = time.Duration
	Location   = time.Location
	Month      = time.Month
	ParseError = time.ParseError
	Time       = time.Time
	Weekday    = time.Weekday

	// the implementation of these types in this package should are used instead of the
	// corresponding standard library types
	// Ticker
	// Timer
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
	Parse           = time.Parse
	ParseDuration   = time.ParseDuration
	ParseInLocation = time.ParseInLocation
	Unix            = time.Unix
	UnixMicro       = time.UnixMicro
	UnixMilli       = time.UnixMilli

	// the following functions have no direct equivalent in this package; equivalent functions
	// are provided as methods of a Clock implementation.  Package-level functions are provided
	// for convenience; these accept a Context parameter which must identify a Context containing
	// a Clock implementation with which to call the function.
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
