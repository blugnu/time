package time

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MockClock extends the Clock interface with methods to manipulate the
// current time of the clock.  In normal use, the underlying clock time
// is advanced only when explicitly directed to do so using AdvanceBy()
// or AdvanceTo() methods; this is "stopped" mode.
//
// When the clock is "running" the current time is advanced semi-automatically
// by the passage of real-time since the last time the clock was updated.
// In "running" mode, the clock is advanced any time that Now() is called,
// or by calling Update().
//
// It is used to simulate the passage of time in tests.
type MockClock interface {
	// MockClock is a mock implementation of the time.Clock interface.
	Clock

	// AdvanceBy moves the current time of the mock clock forward by a
	// specified duration, triggering any timers or tickers that would have
	// been triggered during that passage of time.
	//
	// Calling this method while the clock is running will result in a panic.
	AdvanceBy(d time.Duration)

	// AdvanceTo moves the current time of the mock clock to a specific time,
	// triggering any timers or tickers that would have been triggered during
	// that passage of time.
	//
	// Calling this method while the clock is running will result in a panic.
	AdvanceTo(t time.Time)

	// CreatedAt returns the mocked time at which the clock was started when created.
	CreatedAt() time.Time

	// IsRunning returns true if the clock is in a running state.
	// In this state the clock is advanced by elapsed time whenever Now()
	// is obtained from the clock or when Update() is explicitly called.
	// AdvanceBy() and AdvanceTo() are not supported when the clock is in a
	// running state and will panic.
	//
	// This more closely mimics the behaviour of a real clock but tests using
	// a running clock may be less deterministic and run more slowly than
	// they might.
	IsRunning() bool

	// SinceCreated returns the elapsed mock time since the clock was created.
	// This is the same as calling clock.Since(clock.CreatedAt()).
	SinceCreated() time.Duration

	// Stop stops the clock from advancing automatically.  Every call to
	// Stop() must be matched with a call to Start() to resume automatic
	// advancement.
	//
	// A MockClock is initially created in stopped mode unless the StartRunning
	// option is specified when initialising the clock.
	Stop()

	// Start resumes automatic advancement of the clock.  Every call to
	// Start() must be matched with a call to Stop() to stop automatic
	// advancement.
	//
	// A MockClock is initially created in stopped mode unless the StartRunning
	// option is specified when initialising the clock.  i.e. if the clock
	// is created in stopped mode, an initial call to Start() is required to
	// start the clock.
	Start()

	// Update moves the current time of the mock clock forward by a duration
	// corresponding to the passage of real-time since it was last updated,
	// triggering any timers or tickers that would have been triggered during
	// that passage of time.
	//
	// Calling this method while the clock is stopped will result in a panic.
	Update()
}

// mockClock represents a mock clock that moves forward from an established time and can
// be advanced, rewound or reset at will.
type mockClock struct {
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
		active   tickables
		inactive tickables
	}

	// nextTickerId is the next id to assign to a ticker.
	nextTickerId int
}

// eval is a helper function that executes a supplied function to return a
// value of type T while holding a read lock on a provided clock.
//
// The function must not attempt to acquire a lock on the clock itself, as
// this will result in a deadlock.  The function must also not attempt to
// modify the state of the clock.
func eval[T any](m *mockClock, fn func() T) T {
	m.RLock()
	defer m.RUnlock()
	return fn()
}

func (m *mockClock) panicIfLocked() {
	if !m.TryLock() {
		panic(errClockLocked)
	}
	m.Unlock()
}

func (m *mockClock) withLock(fn func(*mockClock)) {
	m.Lock()
	defer m.Unlock()
	fn(m)
}

// ClockOption represents an option that can be passed to NewMockClock.
type ClockOption func(*mockClock)

// NewMockClock returns an instance of a mock clock.
//
// The default settings on a new clock are:
//
//   - inital time set to the UNIX epoch (00:00:00 UTC on Thursday, 1 Jan 1970)
//   - stopped; advance with AdvanceBy() or AdvanceTo()
//   - does not drop ticks
//   - sleeps the calling goroutine for 1ms on various operations
//
// When stopped, the clock must be explicitly advanced using AdvanceBy() or
// AdvanceTo().  When not stopped Update() may be used to advance the clock
// by the elapsed real-time since the last advancement.
//
// The clock can be customised using the provided options:
//
//   - AtNow() sets the initial time of the mock clock to the current time;
//
//   - AtTime(t time.Time) sets the initial time of the mock clock;
//
//   - DropsTicks() sets the clock to fire tickers only once where multiple
//     ticks would have been triggered by a single advance of the clock
//
//   - WithYield(d time.Duration) sets a duration for which the calling goroutine is
//     suspended before and after each advancement of the clock.
//
//   - StartRunning() sets the mock clock to start in a running state; in the running
//     state the clock is advanced by elapsed time whenever Now() is obtained from
//     the clock or when Update() is explicitly called.  AdvanceBy() and AdvanceTo()
//     are not supported in the running state and will panic.
func NewMockClock(options ...ClockOption) MockClock {
	ret := &mockClock{
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
var _ Clock = (*mockClock)(nil)

// After waits for the duration to elapse and then sends the current time on the returned channel.
func (m *mockClock) After(d time.Duration) <-chan time.Time {
	return m.NewTimer(d).C
}

// AfterFunc waits for the duration to elapse and then executes a function in its own goroutine.
// A Timer is returned that can be stopped.
func (m *mockClock) AfterFunc(d time.Duration, f func()) *Timer {
	return m.newTimer(d, f)
}

// Now returns the current wall time according to the mock clock.
//
// If the clock is frozen, the time will not advance until the clock is unfrozen.
//
// If the clock is not frozen, the clock will first advance by the time elapsed since the
// clock was last updated.
func (m *mockClock) Now() time.Time {
	m.Lock()
	defer m.Unlock()

	return m.advance()
}

// Since returns time since `t` using the mock clock's wall time.
func (m *mockClock) Since(t time.Time) time.Duration {
	return m.Now().Sub(t)
}

// Until returns time until `t` using the mock clock's wall time.
func (m *mockClock) Until(t time.Time) time.Duration {
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
func (m *mockClock) Sleep(d time.Duration) {
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
func (m *mockClock) Tick(d time.Duration) <-chan time.Time {
	if d <= 0 {
		return nil
	}
	return m.NewTicker(d).C
}

// Ticker creates a new instance of Ticker.
func (m *mockClock) NewTicker(d time.Duration) *Ticker {
	return m.newTicker(d)
}

// Timer creates a new Timer.  Since this is a mock implementation, the Timer
// will not fire until the clock is advanced.
func (m *mockClock) NewTimer(d time.Duration) *Timer {
	return m.newTimer(d, nil)
}

// ------------------------------------------------------------------------------------------------

// implements the MockClock interface
var _ MockClock = (*mockClock)(nil)

// advance moves the current time of the mock clock forward by a duration
// corresponding to the passage of real-time since it was last updated.
//
// If the clock is currently stopped the current time is not advanced and must
// be advanced by an explicit interval using AdvanceBy() or AdvanceTo().
//
// This method is not thread-safe and should only be called while the clock
// is locked.
func (m *mockClock) advance() time.Time {
	if !m.IsRunning() {
		return m.now
	}

	var elapsed = time.Since(m.updated)
	m.now = m.now.Add(elapsed)
	m.updated = m.updated.Add(elapsed)

	return m.now
}

// Update moves the current time of the mock clock forward by a duration
// corresponding to the passage of real-time since it was last advanced.
//
// Calling this method while the clock is frozen will result in a panic.
func (m *mockClock) Update() {
	if !m.IsRunning() {
		panic(ErrClockNotRunning)
	}

	m.Lock()
	defer m.Unlock()

	m.advance()
}

// AdvanceBy moves the clock forward by the specified duration.
// This should only be called from a single goroutine at a time.
func (m *mockClock) AdvanceBy(d time.Duration) {
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
func (m *mockClock) AdvanceTo(t time.Time) {
	// a common pattern in tests involving a mock clock is to establish a
	// goroutine to perform some setup or spy, before advancing the mock clock.
	//
	// Yielding here provides room for such goroutines to be established.
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
	m.withLock(func(m *mockClock) {
		m.now = t.In(m.loc)
		m.updated = time.Now()
	})

	// a second yield is provided to allow for any goroutines that are waiting
	// on the clock to be advanced to complete.
	time.Sleep(m.yield)
}

// CreatedAt returns the time at which the clock was created.
func (m *mockClock) CreatedAt() time.Time {
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
func (m *mockClock) IsRunning() bool {
	return m.nStopped.Load() == 0
}

// SinceCreated returns the elapsed mock time since the clock was created.
// This is the same as calling clock.Since(clock.CreatedAt()).
func (m *mockClock) SinceCreated() time.Duration {
	return m.Since(m.CreatedAt())
}

// Start decrements the stop counter on the clock.
func (m *mockClock) Start() {
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
func (m *mockClock) Stop() {
	m.nStopped.Add(1)
}

// ------------------------------------------------------------------------------------------------

func (m *mockClock) resetTicker(t *ticker, d time.Duration) {
	m.withLock(func(m *mockClock) {
		t.d = d
		t.next = m.now.Add(max(d, 0))
	})

	t.enterState(tsActive)
}

func (m *mockClock) resetTimer(t *timer, d time.Duration) {
	m.withLock(func(m *mockClock) {
		if t.next = t.clock.now.Add(d); d == 0 {
			t.tick(t.clock.now)
		}
	})

	if t.state != tsActive {
		t.enterState(tsActive)
	}
}

// activateTicker adds a ticker to the list of active tickers.
func (m *mockClock) activateTicker(t tickable) {
	m.tickers.active = append(m.tickers.active, t)
	sort.Sort(m.tickers.active)
}

// disableTicker moves a ticker from the active list to the inactive list.
func (m *mockClock) disableTicker(id int) {
	var ticker tickable

	if m.tickers.active, ticker = m.tickers.active.take(id); ticker != nil {
		m.tickers.inactive = append(m.tickers.inactive, ticker)
	}
}

// enableTicker moves a ticker from the inactive list to the active list.
func (m *mockClock) enableTicker(id int) {
	var ticker tickable

	if m.tickers.inactive, ticker = m.tickers.inactive.take(id); ticker != nil {
		m.activateTicker(ticker)
	}
}

// newTicker creates a new Ticker backed by a mockTicker.
func (m *mockClock) newTicker(d time.Duration) *Ticker {
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
func (m *mockClock) tick(t time.Time) bool {
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

	m.withLock(func(m *mockClock) {
		sort.Sort(m.tickers.active)
	})

	return true
}

// newTimer creates a new Timer backed by a mocked timer.
func (m *mockClock) newTimer(d time.Duration, fn func()) (result *Timer) {
	m.withLock(func(m *mockClock) {
		// a time.Timer is used to provide a read-only reference to the
		// the channel on which the time is sent when the timer expires
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
