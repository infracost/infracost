terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
    }
    infracost = {
      source = "infracost.io/infracost/infracost"
      version = "0.0.1"
    }
  }
}



provider "aws" {
  region                      = "us-east-1"
  s3_force_path_style         = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

provider "infracost" {}

data "aws_region" "current" {}

variable "aws_ami_id" {
  type    = string
  default = "fake1"
}

module "network" {
  source   = "./network"
  for_each = toset(list("subnet-module-1", "subnet-module-2"))

  subnet_id = each.key
}

resource "aws_network_interface" "root_eip_network_interface" {
  subnet_id   = "subnet-root"
  private_ips = ["10.0.0.1"]
}

resource "aws_eip" "root_nat_eip" {
  network_interface = aws_network_interface.root_eip_network_interface.id
  vpc = true
}

resource "aws_nat_gateway" "root_nat" {
  subnet_id     = "subnet-root"
  allocation_id = aws_eip.root_nat_eip.id
}

resource "infracost_aws_nat_gateway" "root_nat" {
  resources = [aws_nat_gateway.root_nat.id]

  gb_data_processed_monthly {
    value = 100
  }
}

resource "aws_instance" "instance1" {
  ami           = var.aws_ami_id
  instance_type = "t3.micro"

  root_block_device {
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdf"
    volume_size = 10
  }

  ebs_block_device {
    device_name = "xvdg"
    volume_type = "standard"
    volume_size = 20
  }

  ebs_block_device {
    device_name = "xvdh"
    volume_type = "sc1"
    volume_size = 30
  }

  ebs_block_device {
    device_name = "xvdi"
    volume_type = "io1"
    volume_size = 40
    iops        = 1000
  }
}
