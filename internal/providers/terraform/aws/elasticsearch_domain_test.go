package aws_test

import (
	"testing"

	"github.com/infracost/infracost/pkg/testutil"
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestElasticsearchDomain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_elasticsearch_domain" "example" {
		domain_name           = "example"
		elasticsearch_version = "1.5"
	
		cluster_config {
			instance_type = "r4.large.elasticsearch"
		}
	
		snapshot_options {
			automated_snapshot_start_hour = 23
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_elasticsearch_domain.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per instance hour",
					PriceHash:       "198711d36612c57d6f59597e84be3977-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
