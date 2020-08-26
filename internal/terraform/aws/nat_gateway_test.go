package aws_test

import (
	"infracost/internal/testutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNatGatewayIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
resource "aws_nat_gateway" "nat" {
	allocation_id = "eip-12345678"
	subnet_id     = "subnet-12345678"
}`

	resourceCostBreakdowns, err := testutil.RunTFCostBreakdown(tf)
	if err != nil {
		t.Error(err)
	}

	expectedPriceHashes := [][]string{
		{"aws_nat_gateway.nat", "hours", "6e137a9da0718f0ec80fb60866730ba9-d2c98780d7b6e36641b521f1f8145c6f"},
	}

	priceHashResults := testutil.ExtractPriceHashes(resourceCostBreakdowns)

	if !cmp.Equal(priceHashResults, expectedPriceHashes, testutil.PriceHashResultSort) {
		t.Error("got unexpected price hashes", priceHashResults)
	}

	priceComponentCost := testutil.PriceComponentCostFor(resourceCostBreakdowns, "aws_nat_gateway.nat", "hours")
	if !cmp.Equal(priceComponentCost.HourlyCost, priceComponentCost.PriceComponent.Price()) {
		t.Error("got unexpected cost", "aws_nat_gateway.nat", "hours")
	}
}
