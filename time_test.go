package time_test

import (
	"context"
	"sync/atomic"
	"testing"

	. "github.com/blugnu/test"

	"github.com/blugnu/time"
	"github.com/blugnu/time/internal"
)

func TestAfter(t *testing.T) {
	With(t)

	var (
		ctx            = context.Background()
		mockedStart    time.Time
		mockedDuration time.Duration
	)
	ctx, mock := time.ContextWithMockClock(ctx)

	// act: setup a timer that will tick after 10 milliseconds and a waitable
	//      goroutine that stores the time of the first tick when it occurs
	//
	// FUTURE: specify a timeout for the waitable (when supported)

	mockedStart = time.Now(ctx)
	ch := time.After(ctx, 10*time.Millisecond)
	done := internal.Waitable(func() {
		ticked := <-ch
		mockedDuration = ticked.Sub(mockedStart)
	})

	elapsed := time.Dur(func() {
		mock.AdvanceBy(10 * time.Millisecond)
		<-done
	})

	// assert
	Expect(mockedDuration).ToNot(BeLessThan(10 * time.Millisecond))
	Expect(elapsed).To(BeLessThan(8 * time.Millisecond))
}

func TestAfterFunc(t *testing.T) {
	With(t)

	ctx, clock := time.ContextWithMockClock(context.Background())
	var ticked atomic.Bool

	// act
	_ = time.AfterFunc(ctx, 10*time.Millisecond, func() {
		ticked.Store(true)
	})
	clock.AdvanceBy(10 * time.Millisecond)

	// assert
	Expect(ticked.Load()).To(BeTrue())
}

func TestNewTicker(t *testing.T) {
	With(t)

	ctx, clock := time.ContextWithMockClock(context.Background())
	var ticks atomic.Int32

	// act
	ticker := time.NewTicker(ctx, 10*time.Millisecond)
	go func() {
		for range ticker.C {
			ticks.Add(1)
		}
	}()
	clock.AdvanceBy(50 * time.Millisecond)

	// assert
	Expect(ticks.Load()).To(Equal(int32(5)))
}

func TestNewTimer(t *testing.T) {
	With(t)

	ctx, clock := time.ContextWithMockClock(context.Background())
	var ticked atomic.Bool

	// act
	timer := time.NewTimer(ctx, 10*time.Millisecond)
	go func() {
		<-timer.C
		ticked.Store(true)
	}()
	clock.AdvanceBy(10 * time.Millisecond)

	// assert
	Expect(ticked.Load()).To(BeTrue())
}

func TestNow(t *testing.T) {
	With(t)

	tm := time.Date(2023, 10, 1, 2, 3, 4, 5, time.UTC)
	ctx, _ := time.ContextWithMockClock(context.Background(), time.AtTime(tm))

	// act
	now := time.Now(ctx)

	// assert
	Expect(now).To(Equal(tm))
}

func TestSleep(t *testing.T) {
	With(t)

	var (
		ctx, clock = time.ContextWithMockClock(context.Background())
		dur        time.Duration
		sleep      internal.WaitFuncs
	)

	// act
	sleep.Go(func() {
		time.Sleep(ctx, 10*time.Millisecond)
		dur = clock.SinceCreated()
	})
	clock.AdvanceBy(10 * time.Millisecond)
	sleep.Wait()

	// assert
	Expect(dur).To(Equal(10 * time.Millisecond))
}

func TestTick(t *testing.T) {
	With(t)

	ctx, clock := time.ContextWithMockClock(context.Background())
	var ticked atomic.Int32

	// act
	ch := time.Tick(ctx, 10*time.Millisecond)
	go func() {
		for range ch {
			ticked.Add(1)
		}
	}()
	clock.AdvanceBy(50 * time.Millisecond)

	// assert
	Expect(ticked.Load()).To(Equal(int32(5)))
}
