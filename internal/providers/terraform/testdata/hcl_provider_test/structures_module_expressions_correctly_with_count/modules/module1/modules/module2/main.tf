variable "enabled" {
  type    = bool
  default = false
}

resource "aws_ecs_task_definition" "ecs_task" {
  count = var.enabled ? 1 : 0

  requires_compatibilities = ["FARGATE"]
  family                   = "ecs_task_module_2"
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

resource "aws_ecs_service" "ecs_service" {
  count = var.enabled ? 1 : 0

  name            = "ecs_service_module_2"
  launch_type     = "FARGATE"
  task_definition = "${join("", aws_ecs_task_definition.ecs_task.*.family)}:${join("", aws_ecs_task_definition.ecs_task.*.revision)}"
  desired_count   = 1
}
