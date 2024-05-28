provider "aws" {
  region                      = "cn-north-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_lambda_function" "lambda_china" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_china_with_usage" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_china_with_usage_arm" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  architectures = ["arm64"]
}
