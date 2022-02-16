variable "instance_type" {
  default = ""
}

variable "volume_type" {
  default = ""
}
resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.instance_type

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = var.volume_type  # <<<<< Try changing this to gp2 to compare costs
    volume_size = 1000
    iops        = 800
  }
}