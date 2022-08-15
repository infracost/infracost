provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

data "aws_ami" "my_ami" {
  count       = 1
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
  owners = ["099720109477"]
}

locals {
  defaults = {
    instance_type = "m4.large"
    ami           = data.aws_ami.my_ami.*.name[0]
  }
}

resource "aws_instance" "workers_launch_template" {
  instance_type = local.defaults["instance_type"]
  ami           = local.defaults.ami
}
