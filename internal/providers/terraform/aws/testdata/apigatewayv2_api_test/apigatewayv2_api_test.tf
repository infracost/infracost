provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
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