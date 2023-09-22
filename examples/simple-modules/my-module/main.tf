variable "module_volume_type" {
  type    = string
  default = "gp2"
}

variable "module_instance_type" {
  type    = string
  default = "m5.8xlarge"
}

resource "aws_instance" "module_web_app" {
  count         = 2
  ami           = "ami-674cbc1e"
  instance_type = var.module_instance_type

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = var.module_volume_type
    volume_size = 1000
    iops        = 800
  }
}
