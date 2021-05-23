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

resource "aws_config_organization_custom_rule" "my_config_organization_custom_rule" {
  lambda_function_arn = "fake"
  name                = "example"
  trigger_types       = ["ConfigurationItemChangeNotification"]
}

resource "aws_config_organization_custom_rule" "my_config_organization_custom_rule_usage" {
  lambda_function_arn = "fake"
  name                = "example"
  trigger_types       = ["ConfigurationItemChangeNotification"]
}
