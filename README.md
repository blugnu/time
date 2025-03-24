<div align="center" style="margin-bottom:20px">
  <img src=".assets/banner.png" alt="time" />
  <div align="center">
    <a href="https://github.com/blugnu/time/actions/workflows/release.yml">
      <img alt="build-status" src="https://github.com/blugnu/time/actions/workflows/release.yml/badge.svg"/>
    </a>
    <a href="https://goreportcard.com/report/github.com/blugnu/time" >
      <img alt="go report" src="https://goreportcard.com/badge/github.com/blugnu/time"/>
    </a>
    <a>
      <img alt="go version >= 1.14" src="https://img.shields.io/github/go-mod/go-version/blugnu/time?style=flat-square"/>
    </a>
    <a href="https://github.com/blugnu/time/blob/master/LICENSE">
      <img alt="MIT License" src="https://img.shields.io/github/license/blugnu/time?color=%234275f5&style=flat-square"/>
    </a>
    <a href="https://coveralls.io/github/blugnu/time?branch=master">
      <img alt="coverage" src="https://img.shields.io/coveralls/github/blugnu/time?style=flat-square"/>
    </a>
    <a href="https://pkg.go.dev/github.com/blugnu/time">
      <img alt="docs" src="https://pkg.go.dev/badge/github.com/blugnu/time"/>
    </a>
  </div>
</div>

# blugnu/time

A simple and lightweight module for Go providing a context clock and a mock clock for testing time-based
scenarios in accelerated, deterministic time.

## Features

- **Context Clock**: A clock that can be passed in a context;

- **Mock Clock**: A clock that can be used in tests to simulate time passing
  independently of the actual time, allowing for accelerated and deterministic testing of
  time-based code;

- **Mock Timers**: Context deadlines, timeouts, timers and tickers that behave deterministically
  with a mock clock, allowing for testing of time-based code without relying on the system clock or
  the passage of real time;

- **Compatible**: Compatible with the standard library `time` package where appropriate,
  allowing for easy migration from the standard library to `blugnu/time`; provides aliases for
  types and functions that are not clock-dependent with alternative functions for
  clock-dependent functionality;

- **Lightweight**: No external dependencies, making it easy to use and integrate into
  existing projects;

## Installation

```bash
go get github.com/blugnu/time
```

## Usage

As far as possible, `blugnu/time` is designed to be a drop-in replacement for the standard
library `time` package with additional functions where required.

### Clock-Independent Usage

Aliases are provided for constants, types and clock-agnostic functions from the standard
library `time` package:

```golang
      // standard library time
      import "time"
      ux := time.Unix()

      // becomes:
      import "github.com/blugnu/time"
      ux := time.Unix()
```

### Clock-Dependent Functions

Clock-dependent functions are replaced by similes which accept a context, or may be called
using an implementation provided by a clock:

```golang
      // standard library time
      import "time"
      time.Now()
      
      // becomes
      import "github.com/blugnu/time"
      time.Now(ctx)

      // or:
      clock := time.ClockFromContext(ctx)
      clock.Now()
```

Update references to clock-dependent functions to avoid mixing use of mocked and non-mocked
time which would cause unpredictable behaviour in tests.

### Context Deadlines and Timeouts

Context deadlines and timeouts are also clock-dependent.  The `blugnu/time` package provides
a `ContextWithDeadline` and `ContextWithTimeout` functions that return a context with a
deadline or timeout that is based on the clock passed in the context.  When using a mock
clock, the deadline or timeout will be based on the mock clock, allowing for deterministic
behaviour in tests.

```golang
      // standard library time
      import "time"
      ctx, cancel := context.WithDeadline(parentCtx, time.Now().Add(5*time.Second))
      defer cancel()

      // becomes
      import "github.com/blugnu/time"
      ctx, cancel := time.ContextWithDeadline(parentCtx, time.Now().Add(5*time.Second))
      defer cancel()
```

### In Tests

- inject a `MockClock` into the `Context` used for tests;

- use the provided `MockClock` methods to advance the clock in a deterministic fashion to
  exercise time-dependent code, including context deadlines and timeouts independently of the
  elapsed time of the test.

#### Example

```golang
// Simulates testing some code which uses a context with a timeout that would
// normally take 10 seconds to complete if testing the context deadline expiry.
// 
// The test will instead run in milliseconds.
func TestAcceleratedTime() {
  // create a mock clock
  clock := time.NewMockClock()

  // create a context with a 10s timeout; the cancel function is not used 
  timer, _ := clock.ContextWithTimeout(context.Background(), 10*time.Second)

  // start a goroutine that will block until the context is cancelled;
  // a waitgroup is used to sync with the test
  var wg sync.WaitGroup
  wg.Add(1)
  go func() {
    defer wg.Done()
    <-timer.Done()
  }()

  // advance the mock clock by 10s; this will cause the context to be cancelled
  // and the goroutine to unblock
  clock.AdvanceBy(10 * time.Second)
  wg.Wait()

  // verify that the context was cancelled due to the timeout
  if timer.Err() != context.DeadlineExceeded {
    t.Errorf("expected context.DeadlineExceeded, got %s", timer.Err())
  }
}
```

### Additional Functions

Functions are provided for adding or retrieving a clock to/from a context as well as initialising
a mock clock either stand-alone or in a context:

```golang
    // returns the clock from the context or the system clock if not present
    ClockFromContext(ctx context.Context) Clock

    // adds a clock to the context; panics if the context already has a clock
    ContextWithClock(ctx context.Context, clock Clock) context.Context

    // adds a mock clock to the context; panics if the context already has a clock
    ContextWithMockClock(ctx context.Context, opts ...MockClockOption) (context.Context, MockClock)

    // configures a new mock clock
    NewMockClock(opts ...MockClockOption) Clock

    // returns the clock from the context or nil if not present
    TryClockFromContext(ctx context.Context) Clock
```

## System Clock vs Mock Clocks

The system clock is the actual clock of the system, which is used to measure real time.  There
is only one system clock.

A mock clock is a simulated clock that can be used to control the passage of time in tests.
A mock clock can be used to simulate time passing at a different rate than the system clock while
preserving the behaviour of tickers, timers, timeouts and deadlines.

This can provide tests that run more reliably and more quickly than in real-time.  For example,
a test that requires many hours of elapsed clock time can be executed in milliseconds.

Individual tests may use different mock clocks, allowing for different tests to run at different
rates or to simulate different clock behaviours.

## Running vs Stopped Clock

A mock clock can be either running or stopped.  Mock clocks are created stopped by default,
unless the `StartRunning()` option is applied.

### Stopped Clocks

A stopped clock will not advance time automatically.  The clock must be explicitly advanced
using the `AdvanceBy` or `AdvanceTo` methods.  This provides precise control over the passage
of time in tests.

### Running Clocks

When a mock clock is running, it will advance time automatically in real-time whenever
a clock operation is performed involving the current time, or the `Update` method called.

Despite the terminology, a running mock clock will not update in the background.

Attempting to explicitly advance a running clock will result in a panic.  This is to prevent
accidental use of a running clock in tests that expect a stopped clock.

### Stopping and Starting

Although not usually necessary or recommended, a mock clock may be stopped and started using
the `Stop` and `Start` methods.  Every call to `Stop` must be matched with a call to `Start` to
resume running.

Attempting to `Start` a clock that is already running will result in a panic.

## Mock Clock Options

### time.AtNow

The `AtNow` option is a convenience for `time.AtTime(time.SystemClock().Now())`.

### time.AtTime

The `AtTime` option allows you to set the initial time of the mock clock. By default a mock
clock is set to the zero time (Unix Epoch).

### time.DropsTicks

The `DropsTicks` option sets the mock clock to drop any extra ticks when advancing time.
This option only affects tickers, not timers or context deadlines.

By default, if a ticker is set to fire every `1s` and the clock is advanced by `10s`, then
the ticker will fire `10` times, once for each second.  With `DropsTicks` set the ticker will
fire only once in this situation, at the end of the 10 seconds.

### time.InLocation

The `InLocation` option allows you to set the location of the mock clock. The default is UTC.

### time.StartRunning

The `StartRunning` option sets the mock clock to start running immediately when it is
created.  By default, the mock clock is stopped and must be started manually if required.

### time.Yielding

The mock clock suspends the calling goroutine for 1ms when performing certain operations.
The `Yielding` option allows this to be changed to some other duration or disabled entirely
(specifying a duration of 0).
