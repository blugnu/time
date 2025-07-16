package time_test

import (
	"sync/atomic"
	"testing"

	. "github.com/blugnu/test"

	"github.com/blugnu/time"
)

// Tests that AtTime sets the initial time of the mock clock.
func TestClockOption_AtTime(t *testing.T) {
	With(t)

	// arrange
	tm := time.Date(2023, 10, 1, 12, 3, 4, 5, time.UTC)

	// act
	mock := time.NewMockClock(time.AtTime(tm))

	// assert: that the time is set to the expected value
	Expect(mock.Now()).To(Equal(tm))
}

// Tests that AtNow sets the initial time of the mock clock to the current time.
func TestClockOption_AtNow(t *testing.T) {
	With(t)

	// arrange
	sysnow := time.SystemClock().Now()

	// act
	mock := time.NewMockClock(time.AtNow())

	// assert: that the difference between the time reported by the mock
	// and the time immediately before the mock clock was initialised is
	// less than 1ms (it won't be zero since real-time has elapsed)
	Expect(mock.Now().Sub(sysnow)).To(BeLessThan(time.Millisecond))
}

// Tests that DropsTicks sets the mock clock to drop ticks from tickers
func TestClockOption_DropsTicks(t *testing.T) {
	With(t)

	// arrange: establish a mock clock with DropsTicks set, create
	// a ticker to tick every 1s and start a goroutine to count the
	// ticks from the ticker
	var (
		cnt    atomic.Uint32
		clock  = time.NewMockClock(time.DropsTicks())
		ticker = clock.NewTicker(1 * time.Second)
		done   = make(chan struct{})
	)
	defer close(done) // terminates the goroutine we are about to start

	go func() {
		for {
			select {
			case <-ticker.C:
				cnt.Add(1)
			case <-done:
				return
			}
		}
	}()

	// act: advance the clock by 10s
	clock.AdvanceBy(10 * time.Second)

	// assert: the ticker would ordinarily tick 10 times in 10s, but
	// with DropsTicks it should only tick once
	Expect(cnt.Load(), "ticks").To(Equal[uint32](1))
}

// Tests that InLocation sets the location of the mock clock.
func TestClockOption_InLocation(t *testing.T) {
	With(t)

	// arrange
	loc := time.FixedZone("UTC+1", 1*60*60)

	// act
	mock := time.NewMockClock(time.InLocation(loc))

	// assert: that the location is set to the expected value
	Expect(mock.Now().Location()).To(Equal(loc))
}

// Test that StartRunning starts the mock clock in running state.
func TestClockOption_StartRunning(t *testing.T) {
	With(t)

	Run(Test("new mock clocks are not running by default", func() {
		// arrange: create a clock in default (stopped) state
		clock := time.NewMockClock()

		// assert: that the clock is not running
		Expect(clock.IsRunning()).To(BeFalse())
	}))

	Run(Test("StartRunning() option starts the mock clock", func() {
		// arrange: create a clock in default (stopped) state
		clock := time.NewMockClock(time.StartRunning())

		// assert: that the clock IS running
		Expect(clock.IsRunning()).To(BeTrue())
	}))
}

// Tests that YieldTime sets the duration for which the calling goroutine is to be suspended
// when performing operations such as advancing the clock or adding a timer or ticker.
func TestClockOption_YieldTime(t *testing.T) {
	With(t)

	// arrange
	d := 10 * time.Millisecond

	// act: set a larger than usual yield time to make measurement more reliable
	mock := time.NewMockClock(time.YieldTime(d))

	// having set a 10ms Yield time, any advancement of the mock clock will
	// require at least 10ms to complete, even if advanced by a shorter duration

	// assert
	elapsed := time.Dur(func() {
		mock.AdvanceBy(1 * time.Millisecond)
	})
	Expect(elapsed, "elapsed time").ToNot(BeLessThan(d))
}
