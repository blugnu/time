package time

import (
	"testing"
	"time"

	"github.com/blugnu/test"
)

// Tests that AtTime sets the initial time of the mock clock.
func TestClockOption_AtTime(t *testing.T) {
	// arrange
	tm := time.Date(2023, 10, 1, 12, 3, 4, 5, time.UTC)

	// act
	mock := NewMockClock(AtTime(tm))

	// assert: that the time is set to the expected value
	test.Value(t, mock.Now()).Equals(tm)
}

// Tests that AtNow sets the initial time of the mock clock to the current time.
func TestClockOption_AtNow(t *testing.T) {
	// arrange
	tm := time.Now()

	// act
	mock := NewMockClock(AtNow())

	// assert: that the time is set to the expected value
	test.IsTrue(t, mock.Now().Sub(tm) < time.Millisecond)
}

// Tests that InLocation sets the location of the mock clock.
func TestClockOption_InLocation(t *testing.T) {
	// arrange
	loc := time.FixedZone("UTC+1", 1*60*60)

	// act
	mock := NewMockClock(InLocation(loc))

	// assert: that the location is set to the expected value
	test.Value(t, mock.Now().Location()).Equals(loc)
}

// Tests that YieldingFor sets the duration for which the calling goroutine is to be suspended
// when performing operations such as advancing the clock or adding a timer or ticker.
func TestClockOption_YieldingFor(t *testing.T) {
	// arrange
	d := 10 * time.Millisecond

	// act: set a larger than usual yield time to make measurement more reliable
	mock := NewMockClock(Yielding(d))

	// assert: record the time, advance the clock by a period shorter than the yield time
	// and check that the elapsed real time is at least the yield time set on the clock.
	start := time.Now()
	mock.AdvanceBy(1 * time.Millisecond)
	elapsed := time.Since(start)
	test.IsTrue(t, elapsed >= d, "elapsed time")
}
