provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_cloudwatch_log_group" "logs" {
  name = "log-group"
}

resource "aws_cloudwatch_log_group" "logs_withUsage" {
  name = "log-group"
}

resource "aws_cloudwatch_log_group" "logs_count_withUsage" {
  count = 3
  name  = "log-group${count.index}"
}
