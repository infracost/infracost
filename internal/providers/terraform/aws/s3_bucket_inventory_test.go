package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestS3BucketInventoryConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_s3_bucket" "bucket1" {
			bucket = "bucket1"
		}

		resource "aws_s3_bucket" "bucket2" {
			bucket = "bucket2"
		}

		resource "aws_s3_bucket_inventory" "inventory" {
			bucket = aws_s3_bucket.bucket1.bucket
			name = "inventory"
			included_object_versions = "All"

			schedule {
				frequency = "Daily"
			}

			destination {
				bucket {
					format     = "CSV"
					bucket_arn = aws_s3_bucket.bucket2.arn
				}
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_s3_bucket.bucket1",
			SkipCheck: true,
		},
		{
			Name:      "aws_s3_bucket.bucket2",
			SkipCheck: true,
		},
		{
			Name: "aws_s3_bucket_inventory.inventory",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Objects listed",
					PriceHash:        "aa0cc6c33dc5c333d4d8c7333505aadb-262e24dae0e085b444e6d3d16fd79991",
					MonthlyCostCheck: testutil.NilMonthlyCostCheck(),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestS3BucketInventoryConfiguration_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_s3_bucket" "bucket1" {
			bucket = "bucket1"
		}

		resource "aws_s3_bucket" "bucket2" {
			bucket = "bucket2"
		}

		resource "aws_s3_bucket_inventory" "inventory" {
			bucket = aws_s3_bucket.bucket1.bucket
			name = "inventory"
			included_object_versions = "All"

			schedule {
				frequency = "Daily"
			}

			destination {
				bucket {
					format     = "CSV"
					bucket_arn = aws_s3_bucket.bucket2.arn
				}
			}
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_s3_bucket_inventory.inventory": map[string]interface{}{
			"monthly_listed_objects": 10000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name:      "aws_s3_bucket.bucket1",
			SkipCheck: true,
		},
		{
			Name:      "aws_s3_bucket.bucket2",
			SkipCheck: true,
		},
		{
			Name: "aws_s3_bucket_inventory.inventory",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Objects listed",
					PriceHash:        "aa0cc6c33dc5c333d4d8c7333505aadb-262e24dae0e085b444e6d3d16fd79991",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
