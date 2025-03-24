package time

import "sync"

// FUTURE: consider extracting these helpers to a separate module

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
