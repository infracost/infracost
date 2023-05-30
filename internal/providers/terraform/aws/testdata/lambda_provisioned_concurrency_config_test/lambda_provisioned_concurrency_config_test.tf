provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for LambdaProvisionedConcurrencyConfig below
resource "aws_lambda_provisioned_concurrency_config" "with_usage" {
  function_name                     = "lambda_function_name"
  provisioned_concurrent_executions = 50
  qualifier                         = 1
}

resource "aws_lambda_provisioned_concurrency_config" "without_usage" {
  function_name                     = "lambda_function_name"
  provisioned_concurrent_executions = 100
  qualifier                         = 1
}
