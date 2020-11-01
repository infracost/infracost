package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestS3AnalyticsConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_s3_bucket" "bucket1" {
			bucket = "bucket1"
		}

		resource "aws_s3_bucket_analytics_configuration" "bucketanalytics" {
			bucket = aws_s3_bucket.bucket1.bucket
			name   = "bucketanalytics"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_s3_bucket.bucket1",
			SkipCheck: true,
		},
		{
			Name: "aws_s3_bucket_analytics_configuration.bucketanalytics",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Objects monitored",
					PriceHash:        "40e9e08970971a42c21a13af035b210e-262e24dae0e085b444e6d3d16fd79991",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
