package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/pulumi/putest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestS3BucketGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	putest.GoldenFileResourceTests(t, "s3_bucket")
}

func TestS3BucketWithUsage(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Example Pulumi preview JSON with an S3 bucket
	pulumiJSON := `{
		"steps": [{
			"resource": {
				"type": "aws:s3/bucket:Bucket",
				"name": "my-bucket",
				"urn": "urn:pulumi:dev::test::aws:s3/bucket:Bucket::my-bucket",
				"properties": {
					"__defaults": [],
					"bucket": "my-bucket",
					"region": "us-east-1",
					"versioning": {
						"enabled": true
					}
				}
			}
		}]
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"my-bucket": map[string]interface{}{
			"standard": map[string]interface{}{
				"storage_gb":                 1000.0,
				"monthly_tier_1_requests":    100000,
				"monthly_tier_2_requests":    200000,
				"monthly_data_retrieval_gb":  500,
				"monthly_select_data_scanned_gb": 100,
				"monthly_select_data_returned_gb": 50,
			},
			"standard_ia": map[string]interface{}{
				"storage_gb":                 2000.0,
				"monthly_tier_1_requests":    150000,
				"monthly_tier_2_requests":    250000,
				"monthly_data_retrieval_gb":  700,
				"monthly_select_data_scanned_gb": 200,
				"monthly_select_data_returned_gb": 100,
			},
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_s3_bucket.my-bucket",
			CostComponentChecks: []testutil.CostComponentCheck{
				{Name: "Storage (Standard)", MonthlyQuantity: testutil.FloatPtr(1000)},
				{Name: "Storage (Standard - Infrequent Access)", MonthlyQuantity: testutil.FloatPtr(2000)},
				{Name: "GET, SELECT, and all other requests (Standard)", MonthlyQuantity: testutil.FloatPtr(100)},
				{Name: "PUT, COPY, POST, LIST requests (Standard)", MonthlyQuantity: testutil.FloatPtr(200)},
				{Name: "GET, SELECT, and all other requests (Standard - Infrequent Access)", MonthlyQuantity: testutil.FloatPtr(150)},
				{Name: "PUT, COPY, POST, LIST requests (Standard - Infrequent Access)", MonthlyQuantity: testutil.FloatPtr(250)},
				{Name: "Lifecycle transition (Standard to Standard - Infrequent Access)", MonthlyQuantity: testutil.IntPtr(0)},
				{Name: "Data retrieval (Standard - Infrequent Access)", MonthlyQuantity: testutil.FloatPtr(700)},
				{Name: "Select data scanned (Standard)", MonthlyQuantity: testutil.FloatPtr(100)},
				{Name: "Select data returned (Standard)", MonthlyQuantity: testutil.FloatPtr(50)},
				{Name: "Select data scanned (Standard - Infrequent Access)", MonthlyQuantity: testutil.FloatPtr(200)},
				{Name: "Select data returned (Standard - Infrequent Access)", MonthlyQuantity: testutil.FloatPtr(100)},
			},
		},
	}

	putest.ResourceTests(t, pulumiJSON, usage, resourceChecks)
}