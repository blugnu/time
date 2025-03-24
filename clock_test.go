package time

import (
	"sync"
	"testing"
	"time"

	"github.com/blugnu/test"
)

// Ensure that the clock's After channel sends at the correct time.
func TestClock_After(t *testing.T) {
	start := time.Now()
	<-SystemClock().After(20 * time.Millisecond)
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock's AfterFunc executes at the correct time.
func TestClock_AfterFunc(t *testing.T) {
	var ok bool
	var wg sync.WaitGroup

	wg.Add(1)
	start := time.Now()
	SystemClock().AfterFunc(20*time.Millisecond, func() {
		ok = true
		wg.Done()
	})
	wg.Wait()
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
	if !ok {
		t.Fatal("Function did not run")
	}
}

// Ensure that the clock's time matches the standary library.
func TestClock_Now(t *testing.T) {
	a := time.Now().Round(time.Second)
	b := SystemClock().Now().Round(time.Second)
	if !a.Equal(b) {
		t.Errorf("not equal: %s != %s", a, b)
	}
}

func TestClock_Since(t *testing.T) {
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	dur := SystemClock().Since(start)
	test.IsTrue(t, dur >= 10*time.Millisecond)
}

// Ensure that the clock sleeps for the appropriate amount of time.
func TestClock_Sleep(t *testing.T) {
	start := time.Now()
	SystemClock().Sleep(20 * time.Millisecond)
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock ticks correctly.
func TestClock_Tick(t *testing.T) {
	start := time.Now()
	c := SystemClock().Tick(20 * time.Millisecond)
	<-c
	<-c
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 50*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

// Ensure that the clock's ticker ticks correctly.
func TestClock_Ticker(t *testing.T) {
	start := time.Now()
	ticker := SystemClock().Tick(50 * time.Millisecond)
	<-ticker
	<-ticker
	// ticker := RTC().NewTicker(50 * time.Millisecond)
	// <-ticker.C
	// <-ticker.C
	dur := time.Since(start)

	if dur < 100*time.Millisecond || dur > 200*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

func TestClock_Until(t *testing.T) {
	start := time.Now()
	dur := SystemClock().Until(start.Add(10 * time.Millisecond))
	test.IsTrue(t, dur <= 10*time.Millisecond)
}

// Ensure that the clock's ticker can stop correctly.
func TestClock_Ticker_Stop(t *testing.T) {
	ticker := SystemClock().NewTicker(20 * time.Millisecond)
	<-ticker.C
	ticker.Stop()
	select {
	case <-ticker.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's ticker can reset correctly.
func TestClock_Ticker_Reset(t *testing.T) {
	start := time.Now()
	ticker := SystemClock().NewTicker(20 * time.Millisecond)
	<-ticker.C
	ticker.Reset(5 * time.Millisecond)
	<-ticker.C
	dur := time.Since(start)
	if dur >= 30*time.Millisecond {
		t.Fatal("took more than 30ms")
	}
	ticker.Stop()
}

// Ensure that the clock's ticker can stop and then be reset correctly.
func TestClock_Ticker_StopThenReset(t *testing.T) {
	start := time.Now()
	ticker := SystemClock().NewTicker(20 * time.Millisecond)
	<-ticker.C
	ticker.Stop()
	select {
	case <-ticker.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
	ticker.Reset(5 * time.Millisecond)
	<-ticker.C
	dur := time.Since(start)
	if dur >= 60*time.Millisecond {
		t.Fatal("took more than 60ms")
	}
	ticker.Stop()
}

// Ensure that the clock's timer waits correctly.
func TestClock_Timer(t *testing.T) {
	start := time.Now()
	timer := SystemClock().NewTimer(20 * time.Millisecond)
	<-timer.C
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}

	if timer.Stop() {
		t.Fatal("timer still running")
	}
}

// Ensure that the clock's timer can be stopped.
func TestClock_Timer_Stop(t *testing.T) {
	timer := SystemClock().NewTimer(20 * time.Millisecond)
	if !timer.Stop() {
		t.Fatal("timer not running")
	}
	if timer.Stop() {
		t.Fatal("timer wasn't cancelled")
	}
	select {
	case <-timer.C:
		t.Fatal("unexpected send")
	case <-time.After(30 * time.Millisecond):
	}
}

// Ensure that the clock's timer can be reset.
func TestRTCClock_Timer_Reset(t *testing.T) {
	start := SystemClock().Now()
	timer := SystemClock().NewTimer(10 * time.Millisecond)
	if !timer.Reset(20 * time.Millisecond) {
		t.Fatal("timer not running")
	}
	<-timer.C
	dur := time.Since(start)

	if dur < 20*time.Millisecond || dur > 40*time.Millisecond {
		t.Fatalf("Bad duration: %s", dur)
	}
}

func TestRTC_NewTimer_NegativeDuration(t *testing.T) {
	timer := SystemClock().NewTimer(-time.Second)
	select {
	case <-timer.C:
	default:
		t.Fatal("timer should have fired immediately")
	}
}

func TestMockClock_NewTimer_NegativeDuration(t *testing.T) {
	clock := NewMockClock()
	timer := clock.NewTimer(-time.Second)
	select {
	case <-timer.C:
	default:
		t.Fatal("timer should have fired immediately")
	}
}

// Ensure reset can be called immediately after reading channel
func TestClock_Timer_Reset_Unlock(t *testing.T) {
	clock := NewMockClock()
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
