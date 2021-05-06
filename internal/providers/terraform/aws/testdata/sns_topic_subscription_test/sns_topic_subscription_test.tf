provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sns_topic_subscription" "sns_topic_subscription" {
  endpoint = "some-dummy-endpoint"
  protocol = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
}

resource "aws_sns_topic_subscription" "sns_topic_subscription_withUsage" {
  endpoint = "some-dummy-endpoint"
  protocol = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
}