package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestECSService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ecs_cluster" "ecs1" {
			name               = "ecs1"
			capacity_providers = ["FARGATE"]
		}

		resource "aws_ecs_task_definition" "ecs_task1" {
			requires_compatibilities = ["FARGATE"]
			family                   = "ecs_task1"
			memory                   = "2 GB"
			cpu                      = "1 vCPU"
			inference_accelerator {
				device_name = "device1"
				device_type = "eia2.medium"
			}
			container_definitions = <<TASK_DEFINITION
			[
				{
						"command": ["sleep", "10"],
						"entryPoint": ["/"],
						"essential": true,
						"image": "alpine",
						"name": "alpine",
						"network_mode": "none"
				}
			]
			TASK_DEFINITION
		}

		resource "aws_ecs_service" "ecs_fargate1" {
			name            = "ecs_fargate1"
			launch_type     = "FARGATE"
			cluster         = aws_ecs_cluster.ecs1.id
			task_definition = aws_ecs_task_definition.ecs_task1.arn
			desired_count   = 2
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ecs_service.ecs_fargate1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per GB per hour",
					PriceHash:       "8b1ff12686a4c3b2a332da524a724590-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(4)),
				},
				{
					Name:            "Per vCPU per hour",
					PriceHash:       "0c294936ec8abdbfcbb4dfe26cf52afd-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
				{
					Name:            "Inference accelerator (eia2.medium)",
					PriceHash:       "498a3aadc034dfaf873005fdd3f56bbf-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(2)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestECSService_externalDeployment(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_ecs_cluster" "ecs1" {
			name               = "ecs1"
			capacity_providers = ["FARGATE"]
		}

		resource "aws_ecs_service" "ecs_fargate1" {
			name        = "ecs_fargate1"
			launch_type = "FARGATE"
			cluster     = aws_ecs_cluster.ecs1.id

			deployment_controller {
				type = "EXTERNAL"
			}
		}
	`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_ecs_service.ecs_fargate1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Per GB per hour",
					PriceHash:       "8b1ff12686a4c3b2a332da524a724590-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.Zero),
				},
				{
					Name:            "Per vCPU per hour",
					PriceHash:       "0c294936ec8abdbfcbb4dfe26cf52afd-1fb365d8a0bc1f462690ec9d444f380c",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
