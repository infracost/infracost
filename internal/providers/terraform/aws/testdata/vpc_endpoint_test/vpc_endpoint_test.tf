provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_vpc_endpoint" "default" {
  service_name = "com.amazonaws.region.ec2"
  vpc_id       = "vpc-123456"
}

resource "aws_vpc_endpoint" "interface" {
  service_name      = "com.amazonaws.region.ec2"
  vpc_id            = "vpc-123456"
  vpc_endpoint_type = "Interface"
}

resource "aws_vpc_endpoint" "interface_withUsage" {
  service_name      = "com.amazonaws.region.ec2"
  vpc_id            = "vpc-123456"
  vpc_endpoint_type = "Interface"
}

resource "aws_vpc_endpoint" "interface_withBigUsage" {
  service_name      = "com.amazonaws.region.ec2"
  vpc_id            = "vpc-123456"
  vpc_endpoint_type = "Interface"
}

resource "aws_vpc_endpoint" "gateway_loadbalancer" {
  service_name      = "com.amazonaws.region.ec2"
  vpc_id            = "vpc-123456"
  vpc_endpoint_type = "GatewayLoadBalancer"
}

resource "aws_vpc_endpoint" "multiple_interfaces" {
  service_name      = "com.amazonaws.region.ec2"
  vpc_id            = "vpc-123456"
  vpc_endpoint_type = "Interface"
  subnet_ids = [
    "subnet-123456",
    "subnet-654321"
  ]
}

resource "aws_vpc" "main" {
  cidr_block       = "10.0.0.0/16"
  instance_tenancy = "default"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "test2" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.2.0/24"
}

resource "aws_vpc_endpoint" "with_dynamic_subnet" {
  service_name      = "com.amazonaws.region.ec2"
  vpc_id            = aws_vpc.main.id
  vpc_endpoint_type = "Interface"
  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test2.id
  ]
}
