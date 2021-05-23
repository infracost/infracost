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

resource "aws_config_configuration_recorder" "my_config_configuration_recorder" {
  name     = "example"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "r" {
  name = "awsconfig-example"

  assume_role_policy = <<POLICY
{}
POLICY
}

resource "aws_config_configuration_recorder" "my_config_configuration_recorder_usage" {
  name     = "example"
  role_arn = aws_iam_role.rusage.arn
}

resource "aws_iam_role" "rusage" {
  name = "awsconfig-example-usage"

  assume_role_policy = <<POLICY
{}
POLICY
}
