variable "task_cpu" {
  default = 256
}

variable "task_memory" {
  default = 512
}

resource "aws_ecs_task_definition" "task" {
  family                   = "task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  container_definitions = jsonencode([{
    name      = "container"
    image     = "alpine:latest"
    essential = true
  }])
}

resource "aws_ecs_service" "service" {
  name                               = "service"
  task_definition                    = aws_ecs_task_definition.task.arn
  cluster                            = "ecs"
  platform_version                   = "LATEST"
  launch_type                        = "FARGATE"
  propagate_tags                     = "TASK_DEFINITION"
  desired_count                      = 1
  deployment_maximum_percent         = 100
  deployment_minimum_healthy_percent = 0
}
