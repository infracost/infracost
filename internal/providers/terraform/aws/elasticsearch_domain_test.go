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
					Name:            "Instance",
					PriceHash:       "723ac33bae3b8e0751276af954e89a54-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
				{
					Name:            "Storage",
					PriceHash:       "6a8fe5ca25013b67bddcebe1786ad246-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(400)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)

	tfIOEbs := `
	resource "aws_elasticsearch_domain" "example" {
		domain_name           = "example-domain"
		elasticsearch_version = "1.5"

		cluster_config {
			instance_type = "c4.2xlarge.elasticsearch"
			instance_count = 3
		}

		ebs_options {
			ebs_enabled = true
			volume_size = 1000
			volume_type = "io1"
			iops = 10
		}
	}`

	resourceChecksIOEbs := []testutil.ResourceCheck{
		{
			Name: "aws_elasticsearch_domain.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance",
					PriceHash:       "723ac33bae3b8e0751276af954e89a54-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
				{
					Name:            "Storage",
					PriceHash:       "17222df5167b2002292b01078f33d41f-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
				},
				{
					Name:            "Storage IOPS",
					PriceHash:       "cef5d2815d765f1a4d611688519a8cce-9c483347596633f8cf3ab7fdd5502b78",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tfIOEbs, resourceChecksIOEbs)

	tfSTEbs := `
	resource "aws_elasticsearch_domain" "example" {
		domain_name           = "example-domain"
		elasticsearch_version = "1.5"

		cluster_config {
			instance_type = "c4.2xlarge.elasticsearch"
			instance_count = 3
		}

		ebs_options {
			ebs_enabled = true
			volume_size = 123
			volume_type = "standard"
		}
	}`

	resourceChecksSTEbs := []testutil.ResourceCheck{
		{
			Name: "aws_elasticsearch_domain.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance",
					PriceHash:       "723ac33bae3b8e0751276af954e89a54-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(3)),
				},
				{
					Name:            "Storage",
					PriceHash:       "ffa31ac224a19cc7574dbfbffb50722f-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(123)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tfSTEbs, resourceChecksSTEbs)
}
