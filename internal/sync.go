package internal

import (
	"sync"
)

// FUTURE: consider extracting these helpers to a separate module

// Waitable is a helper function that runs the provided function in a goroutine
// and returns a channel that will be closed when the function completes.
// This is useful for waiting on asynchronous operations in tests.
//
// FUTURE: support a timeout parameter to avoid indefinite waits in the event
// that the specified function does not complete.
func Waitable(fn func()) chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		fn()
	}()

	return ch
}

func WaitFor(fn func()) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
	wg.Wait()
}

type WaitFuncs struct {
	sync.WaitGroup
}

func (wg *WaitFuncs) Go(fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

type StartFuncs struct {
	sync.WaitGroup
	init  sync.Once
	start chan struct{}
}

func (wg *StartFuncs) OnStart(fn func()) {
	wg.init.Do(func() {
		wg.start = make(chan struct{})
	})

	// Wait for the start signal before executing the function
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-wg.start
		fn()
	}()
}

func (wg *StartFuncs) Start() {
	close(wg.start)
}
