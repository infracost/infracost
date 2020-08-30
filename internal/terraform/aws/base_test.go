package aws_test

import (
	"infracost/internal/testutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func NewTestIntegration(t *testing.T, r, n, name, priceHash, tf string) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	rn := r + "." + n

	resourceCostBreakdowns, err := testutil.RunTFCostBreakdown(tf)
	if err != nil {
		t.Error(err)
	}

	expectedPriceHashes := [][]string{
		{rn, name, priceHash},
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
