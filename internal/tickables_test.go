package internal //nolint:testpackage // tests rely on non-exported symbols

import (
	"testing"

	. "github.com/blugnu/test"
)

func TestTickerState_String(t *testing.T) {
	With(t)

	tests := []struct {
		state TickerState
		want  string
	}{
		{tsActive, "active"},
		{tsExpired, "expired"},
		{tsStopped, "stopped"},
		{99, "<invalid state(99)>"},
	}

	for _, tt := range tests {
		Run(tt.want, func() {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTickers_Get_NotPresent(t *testing.T) {
	With(t)

	tickers := Tickables{}

	got := tickers.get(1)

	Expect(got).IsNil()
}

func TestTickers_Remove_NotPresent(t *testing.T) {
	With(t)

	tickers := Tickables{
		&timer{tickerId: 1},
		&timer{tickerId: 2},
	}

	got := tickers.remove(3)

	Expect(got).To(DeepEqual(tickers))
}

func TestTickers_Take_NotPresent(t *testing.T) {
	With(t)

	tickers := Tickables{
		&timer{tickerId: 1},
		&timer{tickerId: 2},
	}

	got, ticker := tickers.take(3)

	Expect(got).To(DeepEqual(tickers))
	Expect(ticker).IsNil()
}

func TestTickers_Equal(t *testing.T) {
	With(t)

	Run("equal tickers", func() {
		tickers1 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		tickers2 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		Expect(tickers1.Equal(tickers2)).To(BeTrue())
	})

	Run("not equal tickers", func() {
		tickers1 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		tickers2 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 3},
		}

		Expect(tickers1.Equal(tickers2)).To(BeFalse())
	})

	Run("different lengths", func() {
		tickers1 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		tickers2 := Tickables{
			&timer{tickerId: 1},
		}

		Expect(tickers1.Equal(tickers2)).To(BeFalse())
	})
}
