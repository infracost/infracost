provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ssm_parameter" "ssm_parameter_advanced" {
  name = "my-advanced-ssm-parameter"
  type = "String"
  value = "Advanced Parameter"
  tier = "Advanced"
}

resource "aws_ssm_parameter" "ssm_parameter_advancedWithUsage" {
  name = "my-advanced-ssm-parameter"
  type = "String"
  value = "Advanced Parameter"
  tier = "Advanced"
}
