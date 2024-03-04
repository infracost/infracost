output "region_us2" {
  value = "us-east-2"
}

output "region_eu1" {
  value = "eu-west-1"
}

data "aws_region" "current" {}

resource "aws_instance" "instance_us2" {
  count         = data.aws_region.current.name == "us-east-2" ? 1 : 0
  ami           = "ami-674cbc1e"
  instance_type = "t2.micro"
}

resource "aws_instance" "instance_eu1" {
  count         = data.aws_region.current.name == "eu-west-1" ? 1 : 0
  ami           = "ami-674cbc1e"
  instance_type = "t2.micro"
}
