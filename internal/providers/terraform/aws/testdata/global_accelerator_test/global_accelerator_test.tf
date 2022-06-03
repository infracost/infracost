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

resource "aws_globalaccelerator_accelerator" "fixed_fee" {
  name            = "TestFixedFee"
  ip_address_type = "IPV4"
  enabled         = true
}

resource "aws_globalaccelerator_accelerator" "dt_premium_usage" {
  name            = "TestUsage"
  ip_address_type = "IPV4"
  enabled         = true
}

resource "aws_globalaccelerator_accelerator" "disabled" {
  name            = "TestDisabled"
  ip_address_type = "IPV4"
  enabled         = false
}
