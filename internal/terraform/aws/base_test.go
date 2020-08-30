package aws_test

import (
	"infracost/internal/testutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func NewTestIntegration(r, n, priceHash, tf string) func(t *testing.T) {
	return func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode")
		}

		rn := r + "." + n

		resourceCostBreakdowns, err := testutil.RunTFCostBreakdown(tf)
		if err != nil {
			t.Error(err)
		}

		expectedPriceHashes := [][]string{
			{rn, "hours", priceHash},
		}

		priceHashResults := testutil.ExtractPriceHashes(resourceCostBreakdowns)

		if !cmp.Equal(priceHashResults, expectedPriceHashes, testutil.PriceHashResultSort) {
			t.Error("got unexpected price hashes", priceHashResults)
		}

		priceComponentCost := testutil.PriceComponentCostFor(resourceCostBreakdowns, rn, "hours")
		if !cmp.Equal(priceComponentCost.HourlyCost, priceComponentCost.PriceComponent.Price()) {
			t.Error("got unexpected cost", n, "hours")
		}
	}
}
