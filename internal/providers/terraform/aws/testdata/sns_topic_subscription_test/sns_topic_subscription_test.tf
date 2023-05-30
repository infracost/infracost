provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sns_topic_subscription" "sns_topic_subscription" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
}

resource "aws_sns_topic_subscription" "sns_topic_subscription_withUsage" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "http"
  topic_arn = "arn:aws:sns:us-east-1:123456789123:sns-topic-arn"
}
