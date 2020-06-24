resource "aws_instance" "web_app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  root_block_device {
    volume_size = 15
  }
}

resource "aws_ebs_volume" "storage_option_1" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "io1"
  size              = 15
  iops              = 100
}

resource "aws_ebs_volume" "storage_option_2" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "standard"
  size              = 15
}

resource "aws_eip" "nat_eip" {
  vpc = true
}

resource "aws_nat_gateway" "nat" {
  allocation_id = aws_eip.nat_eip.id
  subnet_id     = tolist(data.aws_subnet_ids.all.ids)[0]
}
