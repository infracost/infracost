package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestApiGatewayStage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
        resource "aws_api_gateway_stage" "cache-1" {
            rest_api_id           = "api-id-1"
            stage_name            = "cache-stage-1"
            deployment_id         = "deployment-id-1"
            cache_cluster_enabled = true
            cache_cluster_size    = 0.5
        }

        resource "aws_api_gateway_stage" "cache-2" {
            rest_api_id           = "api-id-2"
            stage_name            = "cache-stage-2"
            deployment_id         = "deployment-id-2"
            cache_cluster_enabled = true
            cache_cluster_size    = 237
        }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_api_gateway_stage.cache-1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cache memory (0.5 GB)",
					PriceHash:       "13dde350db747bc0cf6b3afb92d76111-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
		{
			Name: "aws_api_gateway_stage.cache-2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Cache memory (237 GB)",
					PriceHash:       "3cbf2c573f7429f90a9acf3a34662d4a-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
