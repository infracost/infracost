provider "aws" {
  region                      = "eu-west-2"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}
resource "aws_kinesis_analytics_application" "withUsage" {
  name = "kinesis-analytics-application-test"
}
resource "aws_kinesis_analytics_application" "withoutUsage" {
  name = "kinesis-analytics-application-test"
}
