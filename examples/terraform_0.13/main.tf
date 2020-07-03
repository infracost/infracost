provider "aws" {
  region = "us-east-1"
}

data "aws_region" "current" {}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnet_ids" "all" {
  vpc_id = data.aws_vpc.default.id
}

data "aws_availability_zones" "available" {
  state = "available"
}

module "web_app" {
  source   = "./web_app"
  for_each = toset(data.aws_availability_zones.available.names)

  availability_zone = each.key
}

module "storage" {
  source   = "./storage"
  count = 3
  depends_on = [module.web_app]
  availability_zone = data.aws_availability_zones.available.names[0]
}
