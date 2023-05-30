provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_dx_gateway_association" "my_aws_dx_gateway_association" {
  dx_gateway_id         = "dx-123456"
  associated_gateway_id = "tgw-123456"
}

resource "aws_dx_gateway_association" "my_aws_dx_gateway_association_usage" {
  dx_gateway_id         = "dx-123456"
  associated_gateway_id = "tgw-123456"
}
