package time

import (
	"testing"

	"github.com/blugnu/test"
)

func TestTickerState_String(t *testing.T) {
	tests := []struct {
		state tickerState
		want  string
	}{
		{tsActive, "active"},
		{tsExpired, "expired"},
		{tsStopped, "stopped"},
		{99, "<invalid state(99)>"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTickers_Get_NotPresent(t *testing.T) {
	tickers := tickables{}

	got := tickers.get(1)

	test.IsNil(t, got)
}

func TestTickers_Remove_NotPresent(t *testing.T) {
	tickers := tickables{
		&timer{tickerId: 1},
		&timer{tickerId: 2},
	}

	got := tickers.remove(3)

	test.Slice(t, got).Equals(tickers)
}

func TestTickers_Take_NotPresent(t *testing.T) {
	tickers := tickables{
		&timer{tickerId: 1},
		&timer{tickerId: 2},
	}

	got, ticker := tickers.take(3)

	test.Slice(t, got).Equals(tickers)
	test.IsNil(t, ticker)
}
