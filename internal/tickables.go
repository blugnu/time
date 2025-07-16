package internal

import (
	"slices"
	"strconv"
	"time"
)

type tickerState int

const (
	tsActive tickerState = iota
	tsExpired
	tsStopped
)

func (state tickerState) String() string {
	switch state {
	case tsActive:
		return "active"
	case tsExpired:
		return "expired"
	case tsStopped:
		return "stopped"
	}
	return "<invalid state(" + strconv.Itoa(int(state)) + ")>"
}

// tickable is an interface that represents a mock timer or ticker.
// It provides methods to get the id, change state, get the next tick time,
// and perform the tick action.
type tickable interface {
	id() int
	enterState(state tickerState)
	nextTick() time.Time
	tick(time.Time) bool
}

// Tickables represents a list of mock Tickables; it supports sorting by
// next tick time.
type Tickables []tickable

func (a Tickables) Len() int           { return len(a) }
func (a Tickables) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Tickables) Less(i, j int) bool { return a[i].nextTick().Before(a[j].nextTick()) }

func (a Tickables) Equal(target Tickables) bool {
	if len(a) != len(target) {
		return false
	}

	a = slices.Clone(a)
	target = slices.Clone(target)

	// Sort both slices by their next tick time
	slices.SortFunc(a, func(t1, t2 tickable) int {
		return t1.nextTick().Compare(t2.nextTick())
	})
	slices.SortFunc(target, func(t1, t2 tickable) int {
		return t1.nextTick().Compare(t2.nextTick())
	})

	for i, ticker := range a {
		if ticker.id() != target[i].id() {
			return false
		}
	}

	return true
}

// get returns the tickable with the given id if present, otherwise returns nil.
func (a Tickables) get(id int) tickable {
	for _, ticker := range a {
		if ticker.id() == id {
			return ticker
		}
	}
	return nil
}

// take if a tickable with the given id is present in the tickers, a new tickers is
// returned with the tickable.  The returned slice has the MockTimer removed.
//
// If there is no tickable with the given id, the original tickers is returned with
// a nil tickable.
func (a Tickables) take(id int) (Tickables, tickable) {
	if ticker := a.get(id); ticker != nil {
		return a.remove(id), ticker
	}
	return a, nil
}

// remove returns a new tickers with the tickable with the given id removed.
func (a Tickables) remove(id int) Tickables {
	for idx, ticker := range a {
		if ticker.id() == id {
			return slices.Delete(a, idx, idx+1)
		}
	}
	return a
}
