provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "instance_type" {
  type    = string
  default = "m5.4xlarge"
}

variable "volume_type" {
  type    = string
  default = "io1"
}

locals {
  instance_type = var.instance_type
  volume_type   = var.volume_type
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = module.my-module[0].aws_instance_type

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = local.volume_type
    volume_size = 1000
    iops        = 800
  }
}

resource "aws_lambda_function" "hello_world" {
  count = 2

  function_name = "hello_world"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  filename      = "function.zip"
  memory_size   = 1024
}

module "my-module" {
  source = "./my-module"
  count  = 2

  module_volume_type   = local.volume_type
  module_instance_type = local.instance_type
}
