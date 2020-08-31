package aws_test

import (
	"testing"

	"infracost/internal/providers/terraform/tftest"
	"infracost/pkg/schema"
	"infracost/pkg/testutil"

	"github.com/shopspring/decimal"
)

func TestEcsService(t *testing.T) {
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

	resources, err := tftest.RunCostCalculation(tf)
	if err != nil {
		t.Error(err)
	}

	expectedPriceHashes := [][]string{
		{"aws_ecs_service.ecs_fargate1", "Per GB per hour", "8b1ff12686a4c3b2a332da524a724590-1fb365d8a0bc1f462690ec9d444f380c"},
		{"aws_ecs_service.ecs_fargate1", "Per vCPU per hour", "0c294936ec8abdbfcbb4dfe26cf52afd-1fb365d8a0bc1f462690ec9d444f380c"},
		{"aws_ecs_service.ecs_fargate1", "Inference accelerator (eia2.medium)", "498a3aadc034dfaf873005fdd3f56bbf-1fb365d8a0bc1f462690ec9d444f380c"},
	}
	testutil.CheckPriceHashes(t, resources, expectedPriceHashes)

	var costComponent *schema.CostComponent

	costComponent = testutil.FindCostComponent(resources, "aws_ecs_service.ecs_fargate1", "Per GB per hour")
	testutil.CheckCost(t, "aws_ecs_service.ecs_fargate1", costComponent, "hourly", costComponent.Price().Mul(decimal.NewFromInt(4)))

	costComponent = testutil.FindCostComponent(resources, "aws_ecs_service.ecs_fargate1", "Per vCPU per hour")
	testutil.CheckCost(t, "aws_ecs_service.ecs_fargate1", costComponent, "hourly", costComponent.Price().Mul(decimal.NewFromInt(2)))

	costComponent = testutil.FindCostComponent(resources, "aws_ecs_service.ecs_fargate1", "Inference accelerator (eia2.medium)")
	testutil.CheckCost(t, "aws_ecs_service.ecs_fargate1", costComponent, "hourly", costComponent.Price().Mul(decimal.NewFromInt(2)))
}
