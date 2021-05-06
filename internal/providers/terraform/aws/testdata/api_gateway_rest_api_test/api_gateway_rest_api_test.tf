provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_api_gateway_rest_api" "api" {
  name              = "rest-api-gateway"
  description       = "Rest API Gateway"
}

resource "aws_api_gateway_rest_api" "my_rest_api" {
  name              = "rest-api-gateway"
  description       = "Rest API Gateway"
}
