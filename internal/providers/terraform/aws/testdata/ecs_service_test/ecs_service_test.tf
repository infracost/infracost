provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

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

resource "aws_ecs_cluster" "ecs2" {
  name               = "ecs2"
  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_service" "ecs_fargate2" {
  name        = "ecs_fargate2"
  launch_type = "FARGATE"
  cluster     = aws_ecs_cluster.ecs2.id

  deployment_controller {
    type = "EXTERNAL"
  }
}

resource "aws_ecs_cluster" "ecs3" {
  name = "ecs1"
}

resource "aws_ecs_service" "ecs_fargate3" {
  name    = "ecs_fargate3"
  cluster = aws_ecs_cluster.ecs3.id

  deployment_controller {
    type = "EXTERNAL"
  }
}
