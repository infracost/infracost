variable "instance_type" {
}

variable "volume_type" {
}

variable "child_instance_type" {
}

output "parent_instance_type" {
  value = aws_instance.web_app.instance_type
}

resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = var.instance_type

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = var.volume_type
    volume_size = 1000
    iops        = 800
  }
}

module "child_instance" {
  source = "./modules/childinstance"

  instance_type = var.child_instance_type
}
