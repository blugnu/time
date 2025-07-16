package internal

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MockClock represents a mock clock that moves forward from an established time and can
// be advanced, rewound or reset at will.
type MockClock struct {
	sync.RWMutex

	// createdAt is the time at which the clock was created.
	// This is used to calculate the elapsed time since the clock was created.
	createdAt time.Time

	// dropsTicks is a flag that when set will cause the mock clock to drop ticks
	// that would have been triggered by tickers; that is if multiple ticks would
	// have been triggered during the passage of time between the last update and
	// the current time, only the last tick will be triggered.
	dropsTicks bool

	// yield is the duration for which the calling goroutine is to be suspended
	// after each time the clock is moved.
	yield time.Duration

	// loc is the location of the clocks mocked time.
	// The default is UTC which may be overridden using the InLocation() option.
	loc *time.Location

	// current time in the location of the clock
	// This is the time that is returned by Now() as used by Since() and Until()
	now time.Time

	// when > 0 the clock will not advance automatically.  Every call to Stop()
	// the clock should be matched by a call to Start().
	nStopped atomic.Int32

	// updated is the last time the clock was queried for the current time
	// (this is the actual, local time according to the system clock)
	//
	// This is used to track elapsed time when advancing the mock clock automatically.
	updated time.Time

	// tickers provides lists of active and inactive tickers.  An inactive ticker
	// is one that has been stopped or has expired (for timers).
	//
	// The active tickers are maintained in order of the next tick time.
	//
	// Maintaining inactive tickers separately allows for tickers to be restarted
	// and for timers to be reset, by returning them to the active list.
	tickers struct {
		active   Tickables
		inactive Tickables
	}

	// nextTickerId is the next id to assign to a ticker.
	nextTickerId int
}

// MockClockOption represents an option that can be passed to NewMockClock.
type MockClockOption func(*MockClock)

// NewMockClock returns a new mock clock
func NewMockClock(options ...MockClockOption) *MockClock {
	ret := &MockClock{
		createdAt: time.Unix(0, 0),
		loc:       time.UTC,
		now:       time.Unix(0, 0).UTC(),
		updated:   time.Now(),
		yield:     1 * time.Millisecond,
	}
	ret.nStopped.Store(1) // start in stopped mode

	for _, opt := range options {
		opt(ret)
	}

	return ret
}

// ------------------------------------------------------------------------------------------------

// implements the Clock interface
var _ Clock = (*MockClock)(nil)

// After waits for the duration to elapse and then sends the current time on the returned channel.
func (m *MockClock) After(d time.Duration) <-chan time.Time {
	return m.NewTimer(d).C
}

// AfterFunc waits for the duration to elapse and then executes a function in its own goroutine.
// A Timer is returned that can be stopped.
func (m *MockClock) AfterFunc(d time.Duration, f func()) *Timer {
	return m.newTimer(d, f)
}

// Now returns the current wall time according to the mock clock.
//
// If the clock is frozen, the time will not advance until the clock is unfrozen.
//
// If the clock is not frozen, the clock will first advance by the time elapsed since the
// clock was last updated.
func (m *MockClock) Now() time.Time {
	m.Lock()
	defer m.Unlock()

	return m.advance()
}

// Since returns time since `t` using the mock clock's wall time.
func (m *MockClock) Since(t time.Time) time.Duration {
	return m.Now().Sub(t)
}

// Until returns time until `t` using the mock clock's wall time.
func (m *MockClock) Until(t time.Time) time.Duration {
	return t.Sub(m.Now())
}

// Sleep pauses the goroutine for the given duration.
//
// If the duration is zero or negative the function returns immediately.
//
// If the clock is running, the duration is passed to time.Sleep() to suspend
// the calling goroutine for the given duration.
//
// If the clock is stopped, the duration is passed to After() and the calling
// goroutine will block until the clock is advanced by at least the specified
// duration.
//
// The clock must be moved forward in a separate goroutine.
func (m *MockClock) Sleep(d time.Duration) {
	if d <= 0 {
		return
	}
	if m.IsRunning() {
		time.Sleep(d)
		return
	}
	<-m.After(max(d, 0))
}

// Tick is a convenience function for Ticker().
// It will return a ticker channel that cannot be stopped or nil if the
// given duration is 0 or negative.
func (m *MockClock) Tick(d time.Duration) <-chan time.Time {
	if d <= 0 {
		return nil
	}
	return m.NewTicker(d).C
}

// Ticker creates a new instance of Ticker.
func (m *MockClock) NewTicker(d time.Duration) *Ticker {
	return m.newTicker(d)
}

// Timer creates a new Timer.  Since this is a mock implementation, the Timer
// will not fire until the clock is advanced.
func (m *MockClock) NewTimer(d time.Duration) *Timer {
	return m.newTimer(d, nil)
}

// ContextWithDeadline returns a new context with the given deadline.
func (m *MockClock) ContextWithDeadline(ctx context.Context, t time.Time) (context.Context, context.CancelFunc) {
	return m.ContextWithDeadlineCause(ctx, t, nil)
}

// ContextWithDeadlineCause returns a new context with the given deadline and cause.
func (m *MockClock) ContextWithDeadlineCause(ctx context.Context, t time.Time, cause error) (context.Context, context.CancelFunc) {
	d := eval(m, func() time.Duration {
		return t.Sub(m.now)
	})
	return m.ContextWithTimeoutCause(ctx, d, cause)
}

// ContextWithTimeout returns a new context with the given timeout.
func (m *MockClock) ContextWithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return m.ContextWithTimeoutCause(ctx, d, nil)
}

// ContextWithTimeoutCause returns a new context with the given timeout and cause.
func (m *MockClock) ContextWithTimeoutCause(ctx context.Context, d time.Duration, cause error) (context.Context, context.CancelFunc) {
	deadline := eval(m, func() time.Time {
		return m.now.Add(d)
	})

	// if the parent context has a deadline which will occur before the timeout
	// return a cancellable context (inheriting the parent deadline since that will
	// be the effective timeout); the specified cause is discarded since the parent
	// will provide the cause of the deadline expiry
	if parentExpires, pd := ctx.Deadline(); pd && parentExpires.Before(deadline) {
		return context.WithCancel(ctx)
	}

	return NewMockContext(ctx, m, deadline, cause)
}

// ------------------------------------------------------------------------------------------------

// Update moves the current time of the mock clock forward by a duration
// corresponding to the passage of real-time since it was last advanced.
//
// Calling this method while the clock is frozen will result in a panic.
func (m *MockClock) Update() {
	if !m.IsRunning() {
		panic(ErrClockNotRunning)
	}

	m.Lock()
	defer m.Unlock()

	m.advance()
}

// AdvanceBy moves the clock forward by the specified duration.
// This should only be called from a single goroutine at a time.
func (m *MockClock) AdvanceBy(d time.Duration) {
	t := eval(m, func() time.Time {
		return m.now.Add(d)
	})
	m.AdvanceTo(t)
}

// AdvanceTo is used to move the current time of the mock clock to a specific time,
// executing all timers that would be triggered during that passage of time.
//
// No attempt is made to simulate the expected elapsed time between the current time
// and the new time or any relative time between timers.
func (m *MockClock) AdvanceTo(t time.Time) {
	// a common pattern in tests involving a mock clock is to establish a
	// goroutine to perform some setup or spy, before advancing the mock clock.
	//
	// yielding here provides room for such goroutines to be established.
	time.Sleep(m.yield)

	// we will only advance the clock to the t if that time is later than the current
	// clock time (the clock cannot be rewound).
	if eval(m, func() bool {
		return m.now.After(t)
	}) {
		panic(ErrNotADelorean)
	}

	// execute timers until there are no more before the new time. If a ticker is
	// ticked, we sort the tickers in case the ticker just ticked now has a new next
	// time later than the previous next ticker next time.
	for m.tick(t) {
	}

	// Ensure that we end with the new time.
	m.withLock(func() {
		m.now = t.In(m.loc)
		m.updated = time.Now()
	})

	// a second yield is provided to allow for any goroutines that are waiting
	// on the clock to be advanced to complete.
	time.Sleep(m.yield)
}

// CreatedAt returns the time at which the clock was created.
func (m *MockClock) CreatedAt() time.Time {
	// this is not mutated after the clock is created so no lock is needed
	return m.createdAt
}

// IsRunning returns true if the clock is in a running state.
//
// In the running state the clock is advanced by elapsed time whenever Now()
// is obtained from the clock or when Update() is explicitly called. AdvanceBy()
// and AdvanceTo() are not supported in the running state and will panic.
//
// The running state is not the default state of the clock; it must be set
// using the StartRunning option when creating the clock or by calling Start()
// on the created clock.
//
// A running clock may be stopped by calling Stop() on that clock.
func (m *MockClock) IsRunning() bool {
	return m.nStopped.Load() == 0
}

// SinceCreated returns the elapsed mock time since the clock was created.
// This is the same as calling clock.Since(clock.CreatedAt()).
func (m *MockClock) SinceCreated() time.Duration {
	return m.Since(m.CreatedAt())
}

// Start decrements the stop counter on the clock.
func (m *MockClock) Start() {
	if n := m.nStopped.Add(-1); n == 0 {
		m.Lock()
		defer m.Unlock()

		m.advance()
	} else if n < 0 {
		panic(ErrClockIsRunning)
	}
}

// Stop increments the stop counter on the clock.  When > 0, the clock
// is prevented from advancing implicitly with Update() and must be advanced
// explicitly using AdvanceBy() or AdvanceTo().
//
// Every call to Stop() must be matched with a call to Start() to resume
// implicit advancement.
func (m *MockClock) Stop() {
	m.nStopped.Add(1)
}

// ------------------------------------------------------------------------------------------------

// advance moves the current time of the mock clock forward by a duration
// corresponding to the passage of real-time since it was last updated.
//
// If the clock is currently stopped the current time is not advanced and must
// be advanced by an explicit interval using AdvanceBy() or AdvanceTo().
//
// This method should only be called when the caller holds the clock's lock.
func (m *MockClock) advance() time.Time {
	if !m.IsRunning() {
		return m.now
	}

	var elapsed = time.Since(m.updated)
	m.now = m.now.Add(elapsed)
	m.updated = m.updated.Add(elapsed)

	return m.now
}

func (m *MockClock) resetTicker(t *ticker, d time.Duration) {
	m.withLock(func() {
		t.d = d
		t.next = m.now.Add(max(d, 0))
	})

	t.enterState(tsActive)
}

func (m *MockClock) resetTimer(t *timer, d time.Duration) {
	m.withLock(func() {
		if t.next = t.clock.now.Add(d); d == 0 {
			t.tick(t.clock.now)
		}
	})

	if t.state != tsActive {
		t.enterState(tsActive)
	}
}

// activateTicker adds a ticker to the list of active tickers.
func (m *MockClock) activateTicker(t tickable) {
	m.tickers.active = append(m.tickers.active, t)
	sort.Sort(m.tickers.active)
}

// disableTicker moves a ticker from the active list to the inactive list.
func (m *MockClock) disableTicker(id int) {
	var ticker tickable

	if m.tickers.active, ticker = m.tickers.active.take(id); ticker != nil {
		m.tickers.inactive = append(m.tickers.inactive, ticker)
	}
}

// enableTicker moves a ticker from the inactive list to the active list.
func (m *MockClock) enableTicker(id int) {
	var ticker tickable

	if m.tickers.inactive, ticker = m.tickers.inactive.take(id); ticker != nil {
		m.activateTicker(ticker)
	}
}

// newTicker creates a new Ticker backed by a mockTicker.
func (m *MockClock) newTicker(d time.Duration) *Ticker {
	m.panicIfLocked()

	ticker := eval(m, func() *Ticker {
		ticker := &Ticker{
			Ticker: &time.Ticker{},
			ticker: &ticker{
				tickerId: m.nextTickerId,
				c:        make(chan time.Time, 1),
				d:        d,
				next:     m.now.Add(max(d, 0)),
				clock:    m,
			},
			initialised: true,
		}
		ticker.C = ticker.c

		m.activateTicker(ticker)
		m.nextTickerId++

		return ticker
	})

	if d <= 0 {
		ticker.tick(m.now)
	}

	return ticker
}

// tick causes the first active ticker before time t (if any) to tick.
// Returns true if a ticker was ticked.
func (m *MockClock) tick(t time.Time) bool {
	m.panicIfLocked()

	ticker := eval(m, func() tickable {
		if len(m.tickers.active) == 0 {
			return nil
		}

		ticker := m.tickers.active[0]
		if ticker.nextTick().After(t) {
			return nil
		}

		return ticker
	})

	if ticker == nil {
		return false
	}

	ticker.tick(t)

	m.withLock(func() {
		sort.Sort(m.tickers.active)
	})

	return true
}

// newTimer creates a new Timer backed by a mocked timer.
func (m *MockClock) newTimer(d time.Duration, fn func()) *Timer {
	var result *Timer

	m.withLock(func() {
		// a time.Timer is used to provide a read-only reference to the
		// channel on which the time is sent when the timer expires
		// (when no function is provided).
		//
		// the time.Timer is not initialised and is not used for timing
		// purposes.
		result = &Timer{
			Timer: &time.Timer{},
			timer: &timer{
				tickerId: m.nextTickerId,
				next:     m.now.Add(max(d, 0)),
				fn:       fn,
				clock:    m,
			},
			initialised: true,
		}

		// if no function is provided, allocate a channel for the timer
		// with a read-only reference in Timer.C
		if fn == nil {
			result.timer.c = make(chan time.Time) // go 1.23+ uses an unbuffered channel
			result.Timer.C = result.timer.c
		}

		m.activateTicker(result)
		m.nextTickerId++
	})

	if d <= 0 {
		result.tick(m.now)
	}

	return result
}

func (m *MockClock) panicIfLocked() {
	if !m.TryLock() {
		panic(ErrClockLocked)
	}

	m.Unlock()
}

func (m *MockClock) withLock(fn func()) {
	m.Lock()
	defer m.Unlock()

	fn()
}

// eval is a helper function that executes a supplied function to return a
// value of type T while holding a read lock on a provided clock.
//
// The function must not attempt to acquire a lock on the clock itself, as
// this will result in a deadlock.  The function must also not attempt to
// modify the state of the clock.
func eval[T any](m *MockClock, fn func() T) T {
	m.RLock()
	defer m.RUnlock()

	return fn()
}
