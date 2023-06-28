provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

provider "aws" {
  alias                       = "eu-west-1"
  region                      = "eu-west-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sns_topic" "sns_topic" {
  name = "my-standard-queue"
}

resource "aws_sns_topic" "sns_topic_another_region" {
  provider = aws.eu-west-1
  name     = "my-standard-queue"
}

resource "aws_sns_topic" "sns_topic_withUsage" {
  name = "my-standard-queue"
}

resource "aws_sns_topic" "sns_topic_withFreeNotifications" {
  name = "my-standard-queue"
}

resource "aws_sns_topic" "sns_topic_withChargedSubscribers" {
  name = "my-standard-queue"
}

resource "aws_sns_topic" "sns_topic_withZeroRequests" {
  name = "my-standard-queue"
}

resource "aws_sns_topic" "sns_topic_customSmsPrice" {
  name = "my-standard-queue"
}

resource "aws_sns_topic" "sns_fifo_topic" {
  name       = "my-fifo-queue.fifo"
  fifo_topic = true
}

resource "aws_sns_topic" "sns_fifo_topic_another_region" {
  provider   = aws.eu-west-1
  name       = "my-fifo-queue.fifo"
  fifo_topic = true
}

resource "aws_sns_topic" "sns_fifo_topic_withUsage" {
  name       = "my-fifo-queue.fifo"
  fifo_topic = true
}

resource "aws_sns_topic" "sns_fifo_topic_withZeroRequests" {
  name       = "my-fifo-queue.fifo"
  fifo_topic = true
}

resource "aws_sns_topic" "sns_fifo_topic_withSubscriptions" {
  name       = "my-fifo-queue.fifo"
  fifo_topic = true
}

resource "aws_sns_topic_subscription" "sns_fifo_topic_subscription1" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "sqs"
  topic_arn = aws_sns_topic.sns_fifo_topic_withSubscriptions.arn
}

resource "aws_sns_topic_subscription" "sns_fifo_topic_subscription2" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "sqs"
  topic_arn = aws_sns_topic.sns_fifo_topic_withSubscriptions.arn
}

resource "aws_sns_topic" "sns_fifo_topic_withUsageAndSubscriptions" {
  name       = "my-fifo-queue.fifo"
  fifo_topic = true
}

resource "aws_sns_topic_subscription" "sns_fifo_topic_subscription3" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "sqs"
  topic_arn = aws_sns_topic.sns_fifo_topic_withUsageAndSubscriptions.arn
}

resource "aws_sns_topic_subscription" "sns_fifo_topic_subscription4" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "sqs"
  topic_arn = aws_sns_topic.sns_fifo_topic_withUsageAndSubscriptions.arn
}

resource "aws_sns_topic_subscription" "sns_fifo_topic_subscription5" {
  endpoint  = "some-dummy-endpoint"
  protocol  = "sqs"
  topic_arn = aws_sns_topic.sns_fifo_topic_withUsageAndSubscriptions.arn
}
