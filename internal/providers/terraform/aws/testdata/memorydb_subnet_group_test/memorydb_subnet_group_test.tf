provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example1" {
  vpc_id     = aws_vpc.example.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "example2" {
  vpc_id     = aws_vpc.example.id
  cidr_block = "10.0.2.0/24"
}

resource "aws_memorydb_subnet_group" "example" {
  name       = "example-subnet-group"
  subnet_ids = [aws_subnet.example1.id, aws_subnet.example2.id]
}
