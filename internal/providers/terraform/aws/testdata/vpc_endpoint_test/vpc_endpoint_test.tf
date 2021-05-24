provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
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
