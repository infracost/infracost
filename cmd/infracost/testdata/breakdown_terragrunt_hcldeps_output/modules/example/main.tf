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

variable "test_input" {
  type = object({})
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
  role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  memory_size   = var.hello_world_function_memory_size
}

// locals blocks to test that dynamic attributes and terraform functions work with the hcl Terragrunt functionality.
locals {
  instances = {
    "m5.4xlarge" = "t2.micro"
  }
}

output "aws_instance_type" {
  value = lookup(local.instances, var.instance_type, "t2.medium")
}
