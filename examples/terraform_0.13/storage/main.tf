variable "availability_zone" {
  type = string
}

resource "aws_ebs_volume" "standard" {
  availability_zone = var.availability_zone
  size              = 30
  type              = "standard"
}

resource "aws_ebs_volume" "io1" {
  availability_zone = var.availability_zone
  type              = "io1"
  size              = 20
  iops              = 00
}

resource "aws_ebs_snapshot" "standard" {
  volume_id = aws_ebs_volume.standard.id
}
