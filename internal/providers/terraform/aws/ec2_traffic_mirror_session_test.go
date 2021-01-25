package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestEC2TrafficMirrorSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ec2_traffic_mirror_session" "session" {
			description = "traffic mirror session"
			network_interface_id = "eni-1234567"
			traffic_mirror_filter_id = "a-traffic-filter-id"
			traffic_mirror_target_id = "a-traffic-target-id"
			session_number = "1"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ec2_traffic_mirror_session.session",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Traffic mirror",
					PriceHash:       "5d359d8dd3efb4d4bc83f7e44c40e1d5-e79b72b3223a1bd297a26b680a122624",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
