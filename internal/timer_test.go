package internal //nolint:testpackage // tests rely on non-exported symbols

import (
	"sync/atomic"
	"testing"
	"time"

	. "github.com/blugnu/test"
)

func TestTimer(t *testing.T) {
	With(t)

	var cnt atomic.Uint32
	clock := SystemClock{}
	timer := clock.NewTimer(1 * time.Millisecond)
	go func() {
		for {
			<-timer.C
			cnt.Add(1)
		}
	}()

	clock.Sleep(600 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](0))

	clock.Sleep(600 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](1))

	active := timer.Reset(5 * time.Millisecond)
	Expect(active, "timer active").To(BeFalse())

	cnt.Store(0)

	clock.Sleep(2600 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](0))

	clock.Sleep(2600 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](1))

	active = timer.Reset(1 * time.Millisecond)
	Expect(active, "timer active").To(BeFalse())

	cnt.Store(0)

	stopped := timer.Stop()
	Expect(stopped, "timer stopped").To(BeTrue())

	clock.Sleep(1500 * time.Microsecond)
	Expect(cnt.Load(), "ticks").To(Equal[uint32](0))
}

func TestTimer_EnterState_InvalidState(t *testing.T) {
	With(t)

	ticker := &timer{}
	defer Expect(Panic(errInvalidState)).DidOccur()

	ticker.enterState(99)
}

func TestTimer_EnterState_NoTransition(t *testing.T) {
	With(t)

	ticker := &timer{state: tsExpired}
	defer Expect(Panic()).DidOccur()

	// act
	ticker.enterState(tsExpired)
}

func TestTimer_Reset_NotInitialized(t *testing.T) {
	With(t)

	timer := &Timer{}
	defer Expect(Panic(errResetCalledOnUninitialized)).DidOccur()

	timer.Reset(time.Second)
}

func TestTimer_Tick_WhenNil(t *testing.T) {
	With(t)

	var sut *timer

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTimer_Tick_WhenNotActive(t *testing.T) {
	With(t)

	var sut = &timer{state: tsExpired}

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	Expect(result).To(BeFalse())
}

func TestTimer_Tick_Premature(t *testing.T) {
	With(t)

	var sut = &timer{
		next: time.Now().Add(+1 * time.Second),
	}

	// act
	result := sut.tick(time.Now())

	// assert: expect false
	Expect(result).To(BeFalse())
}
