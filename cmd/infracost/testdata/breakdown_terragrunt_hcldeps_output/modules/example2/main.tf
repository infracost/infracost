variable "instance_type" {
  description = "The EC2 instance type for the web app"
  type        = string
}

variable "root_block_device_volume_size" {
  description = "The size of the root block device volume for the web app EC2 instance"
  type        = number
}

variable "block_device_volume_size" {
  description = "The size of the block device volume for the web app EC2 instance"
  type        = number
}

variable "block_device_iops" {
  description = "The number of IOPS for the block device for the web app EC2 instance"
  type        = number
}

variable "hello_world_function_memory_size" {
  description = "The memory to allocate to the hello world Lambda function"
  type        = number
}

variable "unspecified_variable" {
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.instance_type

  root_block_device {
    volume_size = var.root_block_device_volume_size
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = var.block_device_volume_size
    iops        = var.block_device_iops
  }
}

resource "aws_lambda_function" "hello_world" {
  function_name = "hello_world"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  memory_size   = var.hello_world_function_memory_size
}

locals {
  bad  = format("%s", var.unspecified_variable.id[0])
  bad2 = format("%s", var.unspecified_variable.id[0])
}

output "bad" {
  value = local.bad
}

output "bad_list" {
  value = [local.bad2]
}

output "good_list_bad_prop" {
  value = [
    {
      id : local.bad
    }
  ]
}

output "bad_map" {
  value = tomap({
    "test" : local.bad
  })
}

output "block_iops" {
  value = 600
}

output "obj" {
  value = {}
}

output "list" {
  value = []
}

output "list_simple" {
  value = []
}

output "list_existing" {
  value = [{ "foo" : "bar" }]
}

output "list_simple_existing" {
  value = ["foo"]
}

output "map" {
  value = tomap({})
}

output "map_existing" {
  value = tomap({ "foo" : "bar" })
}

output "map_simple" {
  value = tomap({})
}

output "improper_mock" {
  value = "invalid-mock-of-a-complex-type"
}

output "improper_mock2" {
  value = "invalid-mock-of-a-complex-type"
}

output "improper_mock3" {
  value = "invalid-mock-of-a-complex-type"
}

output "test_object_type" {
  value = aws_instance.web_app
}
