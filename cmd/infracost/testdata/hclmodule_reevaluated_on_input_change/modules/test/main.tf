variable "enabled" {
  type        = bool
  default     = null
  description = "Set to false to prevent the module from creating any resources"
}


resource "aws_eip" "test" {
  count = var.enabled ? 0 : 1
}

locals {
  instance_type = var.enabled ? "m5.4xlarge" : "t2.micro"
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = local.instance_type

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
