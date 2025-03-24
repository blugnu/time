package time

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestAfterFunc(t *testing.T) {
	ctx, clock := ContextWithMockClock(context.Background())
	var ticked atomic.Bool

	// act
	_ = AfterFunc(ctx, 10*time.Millisecond, func() {
		ticked.Store(true)
	})
	clock.AdvanceBy(10 * time.Millisecond)

	// assert
	test.IsTrue(t, ticked.Load())
}

func TestNewTicker(t *testing.T) {
	ctx, clock := ContextWithMockClock(context.Background())
	var ticks atomic.Int32

	// act
	ticker := NewTicker(ctx, 10*time.Millisecond)
	go func() {
		for range ticker.C {
			ticks.Add(1)
		}
	}()
	clock.AdvanceBy(50 * time.Millisecond)

	// assert
	test.Value(t, ticks.Load()).Equals(5)
}

func TestNewTimer(t *testing.T) {
	ctx, clock := ContextWithMockClock(context.Background())
	var ticked atomic.Bool

	// act
	timer := NewTimer(ctx, 10*time.Millisecond)
	go func() {
		<-timer.C
		ticked.Store(true)
	}()
	clock.AdvanceBy(10 * time.Millisecond)

	// assert
	test.IsTrue(t, ticked.Load())
}

func TestNow(t *testing.T) {
	tm := time.Date(2023, 10, 1, 2, 3, 4, 5, time.UTC)
	ctx, _ := ContextWithMockClock(context.Background(), AtTime(tm))

	// act
	now := Now(ctx)

	// assert
	test.Value(t, now).Equals(tm)
}

func TestSleep(t *testing.T) {
	var (
		ctx, clock = ContextWithMockClock(context.Background())
		dur        time.Duration
		sleep      WaitFuncs
	)

	// act
	sleep.Go(func() {
		Sleep(ctx, 10*time.Millisecond)
		dur = clock.SinceCreated()
	})
	clock.AdvanceBy(10 * time.Millisecond)
	sleep.Wait()

	// assert
	test.Value(t, dur).Equals(10 * time.Millisecond)
}

func TestTick(t *testing.T) {
	ctx, clock := ContextWithMockClock(context.Background())
	var ticked atomic.Int32

	// act
	ch := Tick(ctx, 10*time.Millisecond)
	go func() {
		for range ch {
			ticked.Add(1)
		}
	}()
	clock.AdvanceBy(50 * time.Millisecond)

	// assert
	test.Value(t, ticked.Load()).Equals(5)
}
