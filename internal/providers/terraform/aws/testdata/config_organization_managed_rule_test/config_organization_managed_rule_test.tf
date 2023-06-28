provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_config_organization_managed_rule" "my_config_organization_managed_rule" {
  name            = "example"
  rule_identifier = "IAM_PASSWORD_POLICY"
}

resource "aws_config_organization_managed_rule" "my_config_organization_managed_rule_usage" {
  name            = "example_usage"
  rule_identifier = "IAM_PASSWORD_POLICY"
}
