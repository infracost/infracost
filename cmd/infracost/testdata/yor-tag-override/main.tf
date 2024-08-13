
resource "aws_instance" "web_app" {
  ami           = "ami-674cbc1e"
  instance_type = "m5.4xlarge" # <<<<< Try changing this to m5.8xlarge to compare the costs

  root_block_device {
    volume_size = 50
  }

  ebs_block_device {
    device_name = "my_data"
    volume_type = "io1" # <<<<< Try changing this to gp2 to compare costs
    volume_size = 1000
    iops        = 800
  }

  tags = {
    A = "original-value"
  }
}
