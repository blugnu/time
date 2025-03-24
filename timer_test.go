package time

import (
	"testing"
	"time"

	"github.com/blugnu/test"
)

func TestTimer_EnterState_InvalidState(t *testing.T) {
	ticker := &timer{}
	defer test.ExpectPanic(errInvalidState).Assert(t)

	ticker.enterState(99)
}

func TestTimer_EnterState_NoTransition(t *testing.T) {
	ticker := &timer{state: tsExpired}
	defer test.ExpectPanic(nil).Assert(t)

	// act
	ticker.enterState(tsExpired)
}

func TestTimer_Reset_NotInitialized(t *testing.T) {
	timer := &Timer{}
	defer test.ExpectPanic(errResetCalledOnUninitialized).Assert(t)

	timer.Reset(time.Second)
}

func TestTimer_Tick_WhenNil(t *testing.T) {
	var sut *timer

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	test.IsFalse(t, result)
}

func TestTimer_Tick_WhenNotActive(t *testing.T) {
	var sut = &timer{state: tsExpired}

	// act: attempt to tick a nil timer
	result := sut.tick(time.Time{})

	// assert: expect false
	test.IsFalse(t, result)
}

func TestTimer_Tick_Premature(t *testing.T) {
	var sut = &timer{
		next: time.Now().Add(+1 * time.Second),
	}

	// act
	result := sut.tick(time.Now())

	// assert: expect false
	test.IsFalse(t, result)
}
