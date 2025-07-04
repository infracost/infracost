provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ecs_task_definition" "ecs_task" {
  requires_compatibilities = ["FARGATE"]
  family                   = "ecs_task1"
  memory                   = "2 GB"
  cpu                      = "1 vCPU"

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

resource "aws_ecs_service" "ecs_fargate_no_cluster_1" {
  name            = "ecs_fargate_no_cluster_1"
  launch_type     = "FARGATE"
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 1
}

resource "aws_ecs_service" "ecs_fargate_no_cluster_2" {
  name = "ecs_fargate_no_cluster_2"
  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
    base              = 0
  }
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 2
}

resource "aws_ecs_cluster" "ecs1" {
  name = "ecs1"
}

resource "aws_ecs_cluster_capacity_providers" "cappro1" {
  cluster_name = aws_ecs_cluster.ecs1.name

  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_service" "ecs_fargate1" {
  name            = "ecs_fargate1"
  cluster         = aws_ecs_cluster.ecs1.id
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 1
}

resource "aws_ecs_service" "ecs_fargate11_family" {
  name            = "ecs_fargate1_family"
  launch_type     = "FARGATE"
  task_definition = aws_ecs_task_definition.ecs_task.family
  desired_count   = 1
}

resource "aws_ecs_cluster" "ecs2" {
  name = "ecs2"
}

resource "aws_ecs_cluster_capacity_providers" "cappro2" {
  cluster_name = aws_ecs_cluster.ecs2.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 0
    base              = 1
  }
}


resource "aws_ecs_service" "ecs_fargate2" {
  name            = "ecs_fargate2"
  cluster         = aws_ecs_cluster.ecs2.name
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 2
}

resource "aws_ecs_cluster" "ecs3" {
  name = "ecs3"
}

resource "aws_ecs_cluster_capacity_providers" "cappro3" {
  cluster_name = aws_ecs_cluster.ecs3.name

  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_service" "ecs_fargate3" {
  name            = "ecs_fargate3"
  cluster         = aws_ecs_cluster.ecs3.id
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 3
}

resource "aws_ecs_cluster" "ecs4" {
  name = "ecs4"
}

resource "aws_ecs_cluster_capacity_providers" "cappro4" {
  cluster_name = aws_ecs_cluster.ecs4.name

  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_service" "ecs_fargate4" {
  name            = "ecs_fargate4"
  cluster         = aws_ecs_cluster.ecs4.id
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 4
}

resource "aws_ecs_cluster" "ecs5" {
  name = "ecs4"
}

resource "aws_ecs_cluster_capacity_providers" "cappro5" {
  cluster_name = aws_ecs_cluster.ecs5.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}

resource "aws_ecs_service" "ecs_fargate5" {
  name            = "ecs_fargate5"
  cluster         = aws_ecs_cluster.ecs5.id
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 5
}

resource "aws_ecs_service" "ecs_no_fargate_1" {
  name            = "ecs_no_fargate_1"
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 1
}

resource "aws_ecs_cluster" "ecs_no_fargate_cluster_2" {
  name = "ecs_no_fargate_cluster_2"
}

resource "aws_ecs_service" "ecs_no_fargate_2" {
  name            = "ecs_no_fargate_2"
  cluster         = aws_ecs_cluster.ecs_no_fargate_cluster_2.id
  task_definition = aws_ecs_task_definition.ecs_task.arn
  desired_count   = 2
}

resource "aws_ecs_cluster" "task_set" {
  name = "task-set"
}

resource "aws_ecs_service" "task_set" {
  name          = "task-set"
  launch_type   = "FARGATE"
  desired_count = 2
}

resource "aws_ecs_task_definition" "task_set" {
  requires_compatibilities = ["FARGATE"]
  family                   = "ecs_task1"
  memory                   = "4 GB"
  cpu                      = "2 vCPU"

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

resource "aws_ecs_task_set" "task_set" {
  service         = aws_ecs_service.task_set.id
  cluster         = aws_ecs_cluster.task_set.id
  task_definition = aws_ecs_task_definition.task_set.arn
}
