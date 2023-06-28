provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_apigatewayv2_api" "http" {
  name          = "test-http-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_api" "websocket" {
  name          = "test-websocket-api"
  protocol_type = "WEBSOCKET"
}

resource "aws_apigatewayv2_api" "http_usage" {
  name          = "test-websocket-api"
  protocol_type = "HTTP"
}
resource "aws_apigatewayv2_api" "websocket_usage" {
  name          = "test-websocket-api"
  protocol_type = "WEBSOCKET"
}
