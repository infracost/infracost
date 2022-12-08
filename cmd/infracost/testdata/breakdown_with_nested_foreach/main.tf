provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

locals {
  az = {
    a = {
      az         = "us-east-1a"
      cidr_block = "10.0.16.0/24"
    }
    b = {
      az         = "us-east-1b"
      cidr_block = "10.0.32.0/24"
    }
  }
}

resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
}

resource "aws_subnet" "public" {
  for_each = local.az

  vpc_id                  = aws_vpc.main.id
  cidr_block              = each.value.cidr_block
  availability_zone       = each.value.az
  map_public_ip_on_launch = true

  tags = {
    Name = "public-subnet-${each.value.az}"
  }
}

resource "aws_nat_gateway" "public_nat" {
  for_each      = aws_subnet.public
  subnet_id     = each.value.id
  allocation_id = aws_eip.nat_gateway[each.key].id

  depends_on = [aws_internet_gateway.main]

  tags = {
    Name = "public-nat-${each.value.availability_zone}"
  }
}

resource "aws_eip" "nat_gateway" {
  for_each   = aws_subnet.public
  vpc        = true
  depends_on = [aws_internet_gateway.main]
}
