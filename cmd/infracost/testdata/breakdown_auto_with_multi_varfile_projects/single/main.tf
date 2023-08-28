provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

variable "instance_size" {}
variable "volume_size" {}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.instance_size

  root_block_device {
    volume_size = var.volume_size
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1" # <<<<< Try changing this to gp2 to compare costs
    volume_size = 1000
    iops        = 800
  }
}
