package internal_test

import (
	"testing"

	. "github.com/blugnu/test"

	"github.com/blugnu/time"
	"github.com/blugnu/time/internal"
)

// Tests that AtTime sets the initial time of the mock clock.
func TestClockOption_AtTime(t *testing.T) {
	With(t)

	// arrange
	tm := time.Date(2023, 10, 1, 12, 3, 4, 5, time.UTC)

	// act
	mock := internal.NewMockClock(internal.AtTime(tm))

	// assert: that the time is set to the expected value
	Expect(mock.Now()).To(Equal(tm))
}

// Tests that InLocation sets the location of the mock clock.
func TestClockOption_InLocation(t *testing.T) {
	With(t)

	// arrange
	var (
		loc         = time.FixedZone("UTC+1", 1*60*60)
		initialTime = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	)

	// act
	mock := internal.NewMockClock(
		internal.AtTime(initialTime),
		internal.InLocation(loc),
	)

	// assert: that the location is set to the expected value
	Expect(mock.Now().Location()).To(Equal(loc))
	Expect(mock.Now()).To(Equal(initialTime.In(loc)))
}

// Tests that YieldTime sets the duration for which the calling goroutine is to be suspended
// when performing operations such as advancing the clock or adding a timer or ticker.
func TestClockOption_YieldTime(t *testing.T) {
	With(t)

	// arrange
	d := 10 * time.Millisecond

	// act: set a larger than usual yield time to make measurement more reliable
	mock := internal.NewMockClock(internal.YieldTime(d))

	// assert: record the time, advance the clock by a period shorter than the yield time
	// and check that the elapsed real time is at least the yield time set on the clock.
	elapsed := time.Dur(func() {
		mock.AdvanceBy(1 * time.Millisecond)
	})

	Expect(elapsed).ToNot(BeLessThan(d))
}
