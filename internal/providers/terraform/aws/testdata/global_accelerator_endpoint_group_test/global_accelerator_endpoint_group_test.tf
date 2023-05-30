provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_globalaccelerator_endpoint_group" "eu_west_1" {
  listener_arn          = ""
  endpoint_group_region = "eu-west-1"
}

resource "aws_globalaccelerator_endpoint_group" "us_east_1" {
  listener_arn          = ""
  endpoint_group_region = "us-west-1"
}
