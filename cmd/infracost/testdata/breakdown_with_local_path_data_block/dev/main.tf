provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

data "local_file" "config" {
  filename = "../config/config.json"
}

locals {
  config = jsondecode(data.local_file.config.content)
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = local.config.instance_type

  root_block_device {
    volume_size = 70
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = 3000
    iops        = 1200 # <<<<< Try changing this to 10000 to compare costs
  }
}

resource "aws_lambda_function" "hello_world" {
  function_name = "hello_world"
  role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  memory_size   = 1024 # <<<<< Try changing this to 512 to compare costs
}
