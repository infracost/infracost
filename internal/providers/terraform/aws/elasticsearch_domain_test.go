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
		domain_name           = "example-domain"
		elasticsearch_version = "1.5"
	
		cluster_config {
			instance_type = "c4.2xlarge.elasticsearch"
			instance_count = 3
		}
	
		ebs_options {
			ebs_enabled = true
			volume_size = 400
			volume_type = "gp2"
		}
	}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_elasticsearch_domain.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per instance(x3) hour (example-domain)",
					PriceHash:       "723ac33bae3b8e0751276af954e89a54-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
				{
					Name:            "Storage",
					PriceHash:       "6a8fe5ca25013b67bddcebe1786ad246-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(400)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
