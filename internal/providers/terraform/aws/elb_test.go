package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestELB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_elb" "elb1" {
			listener {
				instance_port     = 80
				instance_protocol = "HTTP"
				lb_port           = 80
				lb_protocol       = "HTTP"
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_elb.elb1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per Classic Load Balancer",
					PriceHash:       "52de45f6e7bf85e2d047a2d9674d9eb2-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
