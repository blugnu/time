package time_test

import (
	"sync"
	"testing"

	. "github.com/blugnu/test"

	"github.com/blugnu/time"
)

func Fatal(msg string) {
	T().Helper()
	T().Fatal(msg)
}

func Fatalf(msg string, args ...any) {
	T().Helper()
	T().Fatalf(msg, args...)
}

func MinMax(mn, mx int, unit time.Duration) (time.Duration, time.Duration) {
	return time.Duration(mn) * unit, time.Duration(mx) * unit
}

// Ensure that the clock's timer can be stopped.
func TestClock_Timer_Stop(t *testing.T) {
	With(t)

	clock := time.SystemClock()
	timer := clock.NewTimer(20 * time.Millisecond)
	if !timer.Stop() {
		Fatal("timer not running")
	}
	if timer.Stop() {
		Fatal("timer wasn't cancelled")
	}
	select {
	case <-timer.C:
		Fatal("unexpected send")
	case <-clock.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's timer can be reset.
func TestSystemClock_Timer_Reset(t *testing.T) {
	With(t)

	dur := time.Dur(func() {
		timer := time.SystemClock().NewTimer(10 * time.Millisecond)
		if !timer.Reset(20 * time.Millisecond) {
			Fatal("timer not running")
		}
		<-timer.C
	})

	Expect(dur).ToNot(BeLessThan(20 * time.Millisecond))
	Expect(dur).ToNot(BeGreaterThan(40 * time.Millisecond))
}

func TestSystemClock_NewTimer_NegativeDuration(t *testing.T) {
	With(t)

	timer := time.SystemClock().NewTimer(-time.Second)
	select {
	case <-timer.C:
	default:
		Fatal("timer should have fired immediately")
	}
}

// Ensure reset can be called immediately after reading channel
func TestClock_Timer_Reset_Unlock(t *testing.T) {
	With(t)

	clock := time.NewMockClock()
	timer := clock.NewTimer(1 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		<-timer.C
		timer.Reset(1 * time.Second)

		<-timer.C
	}()

	clock.AdvanceBy(2 * time.Second)
	wg.Wait()
}
