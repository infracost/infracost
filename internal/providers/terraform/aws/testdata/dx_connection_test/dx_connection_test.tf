provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_dx_connection" "my_dx_connection" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test"
}

resource "aws_dx_connection" "my_dx_connection_usage" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test_Usage"
}

resource "aws_dx_connection" "my_dx_connection_usage_backwards_compat" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test_Usage_Backwards"
}
