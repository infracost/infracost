provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_cloudwatch_event_bus" "my_events" {
  name = "chat-messages"
}

resource "aws_cloudwatch_event_bus" "my_events_withUsage" {
  name = "chat-messages"
}
