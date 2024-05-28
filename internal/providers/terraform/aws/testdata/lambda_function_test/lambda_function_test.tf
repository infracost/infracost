provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_lambda_function" "lambda" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_arm" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  architectures = ["arm64"]
}

resource "aws_lambda_function" "lambda_withUsage" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_withUsage_arm" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  architectures = ["arm64"]
}

resource "aws_lambda_function" "lambda_withUsage512Mem" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 512
}

resource "aws_lambda_function" "lambda_withUsage512Mem_arm" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 512
  architectures = ["arm64"]
}

resource "aws_lambda_function" "lambda_duration_6B" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 512
}

resource "aws_lambda_function" "lambda_duration_75B_arm" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 512
  architectures = ["arm64"]
}

resource "aws_lambda_function" "lambda_duration_9B" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 1024
}

resource "aws_lambda_function" "lambda_duration_11B_arm" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 1024
  architectures = ["arm64"]
}

resource "aws_lambda_function" "lambda_duration_15B" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 2048
}

resource "aws_lambda_function" "lambda_duration_18B_arm" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:aws:resource-id"
  handler       = "exports.test"
  filename      = "function.zip"
  runtime       = "nodejs12.x"
  memory_size   = 2048
  architectures = ["arm64"]
}
