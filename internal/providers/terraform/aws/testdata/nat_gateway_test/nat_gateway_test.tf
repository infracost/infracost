provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_nat_gateway" "nat" {
  allocation_id = "eip-12345678"
  subnet_id     = "subnet-12345678"
}

resource "aws_nat_gateway" "nat_withUsage" {
  allocation_id = "eip-12345678"
  subnet_id     = "subnet-12345678"
}
