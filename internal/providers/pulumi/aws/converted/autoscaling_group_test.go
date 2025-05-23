package aws_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestAutoscalingGroup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	opts := tftest.DefaultGoldenFileOptions()
	opts.CaptureLogs = true
	tftest.GoldenFileResourceTestsWithOpts(t, "autoscaling_group_test", opts)
}

func TestAutoscalingGroup_spot(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	resource "aws_launch_template" "lt_spot" {
		image_id      = "fake_ami"
		instance_type = "t3.medium"
	
		instance_market_options {
			market_type = "spot"
		}
	}
	
	resource "aws_autoscaling_group" "asg_lt_spot" {
		launch_template {
			id = aws_launch_template.lt_spot.id
		}

		desired_capacity = 2
		max_size         = 3
		min_size         = 1
	}
	
	resource "aws_launch_template" "lt_mixed_instance_spot" {
		image_id      = "fake_ami"
		instance_type = "t3.medium"
	}
	
	resource "aws_autoscaling_group" "asg_mixed_instance_spot" {
		desired_capacity = 6
		max_size         = 10
		min_size         = 1
	
		mixed_instances_policy {
			launch_template {
				launch_template_specification {
					launch_template_id = aws_launch_template.lt_mixed_instance_spot.id
				}
			}
	
			instances_distribution {
				on_demand_base_capacity                  = 2
				on_demand_percentage_above_base_capacity = 50
			}
		}
	}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_autoscaling_group.asg_lt_spot",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.lt_spot",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Instance usage (Linux/UNIX, spot, t3.medium)",
							PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-803d7f1cd2f621429b63f791730e7935",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
						},
						{
							Name:             "CPU credits",
							PriceHash:        "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
						},
					},
				},
			},
		},
		{
			Name: "aws_autoscaling_group.asg_mixed_instance_spot",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.lt_mixed_instance_spot",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Instance usage (Linux/UNIX, on-demand, t3.medium)",
							PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
						},
						{
							Name:            "Instance usage (Linux/UNIX, spot, t3.medium)",
							PriceHash:       "c8faba8210cd512ccab6b71ca400f4de-803d7f1cd2f621429b63f791730e7935",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
						},
						{
							Name:             "CPU credits",
							PriceHash:        "ccdf11d8e4c0267d78a19b6663a566c1-e8e892be2fbd1c8f42fd6761ad8977d8",
							MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.UsageMap{}, resourceChecks)
}
