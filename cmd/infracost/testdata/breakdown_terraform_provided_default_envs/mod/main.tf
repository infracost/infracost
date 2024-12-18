provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "env" {}

locals {
  instance_types = {
    dev  = "t2.micro"
    prod = "t3.medium"
    stg  = "t3.large"
  }
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = local.instance_types[var.env]

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1"
    volume_size = 1000
    iops        = 800
  }
}
