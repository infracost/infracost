package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestLightsailInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_lightsail_instance" "linux1" {
			name              = "linux1"
			availability_zone = "us-east-1a"
			blueprint_id      = "centos_7_1901_01"
			bundle_id         = "xlarge_2_0"
		}

		resource "aws_lightsail_instance" "win1" {
			name              = "win1"
			availability_zone = "us-east-1a"
			blueprint_id      = "windows_2019"
			bundle_id         = "small_win_2_0"
		}		`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_lightsail_instance.linux1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Virtual server (Linux/UNIX)",
					PriceHash:       "d1a975f1c2f812954c1e0d1b25c15117-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_lightsail_instance.win1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Virtual server (Windows)",
					PriceHash:       "0163422d7cc913ae205cb3626e3d98b2-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
