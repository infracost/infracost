package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestRoute53Zone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_route53_zone" "zone1" {
			name = "example.com"
		}
		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_route53_zone.zone1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Hosted zone",
					PriceHash:       "11a88b17c107a718b150e048d21ce5ac-48bca87a3e73bd3aa593065935882019",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
