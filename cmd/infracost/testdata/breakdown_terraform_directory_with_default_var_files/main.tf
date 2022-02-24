provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.instance_type

  root_block_device {
    volume_size = var.root_block_size
  }

  ebs_block_device {
    device_name = "block1"
    volume_type = var.block1_volume_type
    volume_size = var.block1_ebs_volume_size
    iops        = var.block1_ebs_iops
  }

  ebs_block_device {
    device_name = "block2"
    volume_type = var.block2_volume_type
    volume_size = var.block2_ebs_volume_size
    iops        = var.block2_ebs_iops
  }
}

variable "instance_type" {
  default = "m5.4xlarge"
}

variable "root_block_size" {
  type = number
}

variable "block1_volume_type" {
  type = string
}

variable "block1_ebs_volume_size" {
  type = number
}

variable "block1_ebs_iops" {
  type = number
}

variable "block2_volume_type" {
  type = string
}

variable "block2_ebs_volume_size" {
  type = number
}

variable "block2_ebs_iops" {
  type = number
}
