data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_eip" "available" {
  for_each = { for entry in data.aws_availability_zones.available.names : "${entry}" => entry }
}

resource "aws_eip" "current" {
  for_each = { for entry in [data.aws_region.current] : "${entry.name}" => entry }
}
