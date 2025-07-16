package internal //nolint:testpackage // tests rely on non-exported symbols

import (
	"testing"

	. "github.com/blugnu/test"
)

func TestTickerState_String(t *testing.T) {
	With(t)

	type testcase struct {
		state tickerState
		want  string
	}
	Run(Testcases(
		ForEach(func(tc testcase) {
			got := tc.state.String()
			Expect(got).To(Equal(tc.want))
		}),
		Cases([]testcase{
			{tsActive, "active"},
			{tsExpired, "expired"},
			{tsStopped, "stopped"},
			{99, "<invalid state(99)>"},
		})))
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

	Run(Test("equal tickers", func() {
		tickers1 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		tickers2 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		Expect(tickers1.Equal(tickers2)).To(BeTrue())
	}))

	Run(Test("not equal tickers", func() {
		tickers1 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		tickers2 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 3},
		}

		Expect(tickers1.Equal(tickers2)).To(BeFalse())
	}))

	Run(Test("different lengths", func() {
		tickers1 := Tickables{
			&timer{tickerId: 1},
			&timer{tickerId: 2},
		}

		tickers2 := Tickables{
			&timer{tickerId: 1},
		}

		Expect(tickers1.Equal(tickers2)).To(BeFalse())
	}))
}
