package aws_test

import (
	"testing"

	"github.com/infracost/infracost/pkg/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestAutoscalingGroup_launchConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_launch_configuration" "lc1" {
			image_id      = "fake_ami"
			instance_type = "t2.medium"

			root_block_device {
				volume_size = 10
			}

			ebs_block_device {
				device_name = "xvdf"
				volume_size = 10
			}
		}

		resource "aws_autoscaling_group" "asg1" {
			launch_configuration = aws_launch_configuration.lc1.id
			desired_capacity     = 2
			max_size             = 3
			min_size             = 1
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_autoscaling_group.asg1",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_configuration.lc1",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Linux/UNIX usage (on-demand, t2.medium)",
							PriceHash:       "250382a8c0da495d6048e6fc57e526bc-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
						},
					},
					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "root_block_device",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
								},
							},
						},
						{
							Name: "ebs_block_device[0]",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
								},
							},
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}

func TestAutoscalingGroup_launchTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_launch_template" "lt1" {
			image_id      = "fake_ami"
			instance_type = "t2.medium"

			block_device_mappings {
				device_name = "xvdf"
				ebs {
					volume_size = 10
				}
			}

			block_device_mappings {
				device_name = "xvfa"
				ebs {
					volume_size = 20
					volume_type = "io1"
					iops        = 200
				}
			}
		}

		resource "aws_autoscaling_group" "asg1" {
			desired_capacity = 2
			max_size         = 3
			min_size         = 1

			launch_template {
				id = aws_launch_template.lt1.id
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_autoscaling_group.asg1",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.lt1",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Linux/UNIX usage (on-demand, t2.medium)",
							PriceHash:       "250382a8c0da495d6048e6fc57e526bc-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
						},
					},
					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "root_block_device",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(16)),
								},
							},
						},
						{
							Name: "block_device_mapping[0]",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
								},
							},
						},
						{
							Name: "block_device_mapping[1]",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:            "Storage",
									PriceHash:       "99450513de8c131ee2151e1b319d8143-ee3dd7e4624338037ca6fea0933a662f",
									HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
								},
								{
									Name:            "Storage IOPS",
									PriceHash:       "d5c5e1fb9b8ded55c336f6ae87aa2c3b-9c483347596633f8cf3ab7fdd5502b78",
									HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(400)),
								},
							},
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}

func TestAutoscalingGroup_mixedInstanceLaunchTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_launch_template" "lt1" {
			image_id      = "fake_ami"
			instance_type = "t2.medium"
		}

		resource "aws_autoscaling_group" "asg1" {
			desired_capacity   = 6
			max_size           = 10
			min_size           = 1

			mixed_instances_policy {
				launch_template {
					launch_template_specification {
						launch_template_id = aws_launch_template.lt1.id
					}

					override {
						instance_type     = "t2.large"
						weighted_capacity = "2"
					}

					override {
						instance_type     = "t2.xlarge"
						weighted_capacity = "4"
					}
				}

				instances_distribution {
					on_demand_base_capacity = 1
					on_demand_percentage_above_base_capacity = 50
				}
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_autoscaling_group.asg1",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.lt1",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Linux/UNIX usage (on-demand, t2.large)",
							PriceHash:       "3aa92af51438c0eba9dc1c62539adf5b-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
						},
						{
							Name:            "Linux/UNIX usage (spot, t2.large)",
							PriceHash:       "3aa92af51438c0eba9dc1c62539adf5b-803d7f1cd2f621429b63f791730e7935",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
						},
					},
					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "root_block_device",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(24)),
								},
							},
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}

func TestAutoscalingGroup_mixedInstanceLaunchTemplateDynamic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_launch_template" "lt1" {
			image_id      = "fake_ami"
			instance_type = "t2.medium"
		}

		resource "aws_autoscaling_group" "asg1" {
			desired_capacity   = 3
			max_size           = 5
			min_size           = 1

			mixed_instances_policy {
				launch_template {
					launch_template_specification {
						launch_template_id = aws_launch_template.lt1.id
					}

					dynamic "override" {
						for_each = ["t2.large", "t2.xlarge"]

						content {
							instance_type = override.value
						}
					}
				}

				instances_distribution {
					on_demand_base_capacity = 1
					on_demand_percentage_above_base_capacity = 50
				}
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_autoscaling_group.asg1",
			SubResourceChecks: []testutil.ResourceCheck{
				{
					Name: "aws_launch_template.lt1",
					CostComponentChecks: []testutil.CostComponentCheck{
						{
							Name:            "Linux/UNIX usage (on-demand, t2.large)",
							PriceHash:       "3aa92af51438c0eba9dc1c62539adf5b-d2c98780d7b6e36641b521f1f8145c6f",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
						},
						{
							Name:            "Linux/UNIX usage (spot, t2.large)",
							PriceHash:       "3aa92af51438c0eba9dc1c62539adf5b-803d7f1cd2f621429b63f791730e7935",
							HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
						},
					},
					SubResourceChecks: []testutil.ResourceCheck{
						{
							Name: "root_block_device",
							CostComponentChecks: []testutil.CostComponentCheck{
								{
									Name:             "Storage",
									PriceHash:        "efa8e70ebe004d2e9527fd30d50d09b2-ee3dd7e4624338037ca6fea0933a662f",
									MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(24)),
								},
							},
						},
					},
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
