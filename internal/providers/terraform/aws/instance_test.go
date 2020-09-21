package aws_test

import (
	"testing"

	"github.com/infracost/infracost/pkg/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_instance" "instance1" {
			ami           = "fake_ami"
			instance_type = "m3.medium"

			root_block_device {
				volume_size = 10
			}

			ebs_block_device {
				device_name = "xvdf"
				volume_size = 10
			}

			ebs_block_device {
				device_name = "xvdg"
				volume_type = "standard"
				volume_size = 20
			}

			ebs_block_device {
				device_name = "xvdh"
				volume_type = "sc1"
				volume_size = 30
			}

			ebs_block_device {
				device_name = "xvdi"
				volume_type = "io1"
				volume_size = 40
				iops        = 1000
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.instance1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (on-demand, m3.medium)",
					PriceHash:       "666e02bbe686f6950fd8a47a55e83a75-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "root_block_device",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage",
							PriceHash:       "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
						},
					},
				},
				{
					Name: "ebs_block_device[0]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage",
							PriceHash:       "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(10)),
						},
					},
				},
				{
					Name: "ebs_block_device[1]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage",
							PriceHash:       "0ed17ed1777b7be91f5b5ce79916d8d8-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
						},
					},
				},
				{
					Name: "ebs_block_device[2]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage",
							PriceHash:       "3122df29367c2460c76537cccf0eadb5-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
						},
					},
				},
				{
					Name: "ebs_block_device[3]",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Storage",
							PriceHash:       "99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
						},
						{
							Name:            "Storage IOPS",
							PriceHash:       "d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78",
							HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1000)),
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}

func TestInstance_ebsOptimized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_instance" "instance1" {
			ami           = "fake_ami"
			instance_type = "m3.large"
			ebs_optimized = true
		}

		resource "aws_instance" "instance2" {
			ami           = "fake_ami"
			instance_type = "r3.xlarge"
			ebs_optimized = true
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_instance.instance1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (on-demand, m3.large)",
					PriceHash:       "1abac89a8296443758727a2728579a2a-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
		{
			Name: "aws_instance.instance2",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Compute (on-demand, r3.xlarge)",
					PriceHash:       "5fc0daede99fac3cce64d575979d7233-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:            "EBS-Optimized Usage",
					PriceHash:       "7f4fb9da921a628aedfbe150d930e255-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name:      "root_block_device",
					SkipCheck: true,
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
