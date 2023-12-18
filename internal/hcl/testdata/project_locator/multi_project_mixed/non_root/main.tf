resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge"

  root_block_device {
    volume_size = 50
  }
}

