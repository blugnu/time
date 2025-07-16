package internal_test

import (
	"context"
	"testing"

	. "github.com/blugnu/test"

	"github.com/blugnu/time/internal"
)

// Tests that FromContext returns the clock present in the context.
func TestContextClock(t *testing.T) {
	With(t)

	ctx := context.Background()

	Run(Test("with no clock in context", func() {
		// act
		clock := internal.ClockFromContext(ctx)

		// assert
		Expect(clock).IsNil()
	}))

	Run(Test("with clock in context", func() {
		// arrange
		mock := internal.NewMockClock()
		ctx := internal.ContextWithClock(ctx, mock)

		// act
		clock := internal.ClockFromContext(ctx)

		// assert
		Require(clock).IsNotNil()
		ExpectType[*internal.MockClock](clock)
	}))
}
