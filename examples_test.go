package time_test

import (
	"fmt"
	"math"
	"sync/atomic"

	"github.com/blugnu/time"
)

func ExampleMockClock_After() {
	// create a new mock clock and initialise an atomic bool which
	// will be used to identify when the timer has fired
	var (
		clock    = time.NewMockClock()
		fired    atomic.Bool
		realtime = time.SystemClock()
		start    = realtime.Now()
	)

	// establish a timer to run after 10 seconds and start a goroutine
	// that will set fired when the timer is executed
	ch := clock.After(10 * time.Second)
	go func() {
		<-ch
		fired.Store(true)
	}()

	// Print the starting value.
	fmt.Printf("%s: timer fired: %v\n", clock.Now(), fired.Load())

	// Move the clock forward 5 seconds and print the value again.
	clock.AdvanceBy(5 * time.Second)
	fmt.Printf("%s: timer fired: %v\n", clock.Now(), fired.Load())

	// Move the clock forward 5 seconds to the tick time and check the value.
	clock.AdvanceBy(5 * time.Second)
	fmt.Printf("%s: timer fired: %v\n", clock.Now(), fired.Load())

	// Check if the elapsed time is less than 200ms.
	if realtime.Since(start) < 200*time.Millisecond {
		fmt.Println("completed in < 200ms")
	}

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: timer fired: false
	// 1970-01-01 00:00:05 +0000 UTC: timer fired: false
	// 1970-01-01 00:00:10 +0000 UTC: timer fired: true
	// completed in < 200ms
}

func ExampleMockClock_AfterFunc() {
	// Create a new mock clock.
	clock := time.NewMockClock()
	var cnt atomic.Uint32
	cnt.Add(1)

	// Execute a function after 10 mock seconds.
	clock.AfterFunc(10*time.Second, func() {
		cnt.Add(1)
	})

	// Print the starting value.
	fmt.Printf("%s: %d\n", clock.Now().UTC(), cnt.Load())

	// Move the clock forward 10 seconds and print the new value.
	clock.AdvanceBy(10 * time.Second)
	fmt.Printf("%s: %d\n", clock.Now().UTC(), cnt.Load())

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: 1
	// 1970-01-01 00:00:10 +0000 UTC: 2
}

func ExampleMockClock_Sleep() {
	// Create a new mock clock.
	var (
		clock    = time.NewMockClock()
		sleeping atomic.Bool
	)

	// Execute a function after 10 mock seconds.
	sleeping.Store(true)
	go func() {
		clock.Sleep(10 * time.Second)
		sleeping.Store(false)
	}()

	// Print the starting value.
	fmt.Printf("%s: sleeping: %v\n", clock.Now(), sleeping.Load())

	// Move the clock forward 10 seconds and print the new value.
	clock.AdvanceBy(10 * time.Second)
	fmt.Printf("%s: sleeping: %v\n", clock.Now(), sleeping.Load())

	// Output:
	// 1970-01-01 00:00:00 +0000 UTC: sleeping: true
	// 1970-01-01 00:00:10 +0000 UTC: sleeping: false
}

func ExampleMockClock_NewTicker() {
	// Create a new mock clock.
	clock := time.NewMockClock()
	var cnt atomic.Uint32

	ready := make(chan struct{})
	// Increment count every mock second.
	go func() {
		ticker := clock.NewTicker(1 * time.Second)
		close(ready)
		for {
			<-ticker.C
			cnt.Add(1)
		}
	}()
	<-ready

	// Move the clock forward 10 seconds and print the new value.
	clock.AdvanceBy(10 * time.Second)
	fmt.Printf("Count is %d after 10 seconds\n", cnt.Load())

	// Move the clock forward 5 more seconds and print the new value.
	clock.AdvanceBy(5 * time.Second)
	fmt.Printf("Count is %d after 15 seconds\n", cnt.Load())

	// Output:
	// Count is 10 after 10 seconds
	// Count is 15 after 15 seconds
}

func ExampleMockClock_NewTimer() {
	// Create a new mock clock.
	clock := time.NewMockClock()
	var cnt atomic.Uint32

	ready := make(chan struct{})
	// Increment count after a mock second.
	go func() {
		timer := clock.NewTimer(1 * time.Second)
		close(ready)
		<-timer.C
		cnt.Add(1)
	}()
	<-ready

	// Move the clock forward 10 seconds and print the new value.
	clock.AdvanceBy(10 * time.Second)
	fmt.Printf("Count is %d after 10 seconds\n", cnt.Load())

	// Output:
	// Count is 1 after 10 seconds
}

func ExampleTicker() {
	// using the system clock
	clock := time.SystemClock()
	start := clock.Now()

	// create a ticker that ticks every 100ms
	ticker := clock.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// output the number and time (since starting) of each tick
	n := 0
	go func() {
		for range ticker.C {
			n++
			fmt.Printf("fire %d @ ~%dms\n", n, 10*int(math.Trunc(float64(clock.Since(start).Milliseconds())/10)))
		}
	}()

	// wait for a little over 500ms
	clock.Sleep(510 * time.Millisecond)

	// Output:
	// fire 1 @ ~100ms
	// fire 2 @ ~200ms
	// fire 3 @ ~300ms
	// fire 4 @ ~400ms
	// fire 5 @ ~500ms
}
