provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ssm_activation" "ssm_activation" {
  name               = "test_ssm_advanced_activation"
  description        = "Test"
  iam_role           = "my-test-iam-role"
  registration_limit = 1001
}

resource "aws_ssm_activation" "ssm_activation_withUsage" {
  name        = "test_ssm_advanced_activation"
  description = "Test"
  iam_role    = "my-test-iam-role"
}
