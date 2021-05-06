provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sns_topic" "sns_topic" {
  name = "my-standard-queue"
}

resource "aws_sns_topic" "sns_topic_withUsage" {
  name = "my-standard-queue"
}