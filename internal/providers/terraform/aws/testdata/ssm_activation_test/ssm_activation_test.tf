provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ssm_activation" "ssm_activation" {
  name = "test_ssm_advanced_activation"
  description = "Test"
  iam_role = "my-test-iam-role"
  registration_limit = 1001
}

resource "aws_ssm_activation" "ssm_activation_withUsage" {
  name = "test_ssm_advanced_activation"
  description = "Test"
  iam_role = "my-test-iam-role"
}
