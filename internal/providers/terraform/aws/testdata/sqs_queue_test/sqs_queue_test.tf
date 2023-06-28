provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_sqs_queue" "standard_sqs_queue" {
  name       = "my-standard-queue"
  fifo_queue = false
}

resource "aws_sqs_queue" "fifo_sqs_queue" {
  name       = "my.fifo"
  fifo_queue = true
}

resource "aws_sqs_queue" "standard_sqs_queue_withUsage" {
  name       = "my-standard-queue"
  fifo_queue = false
}

resource "aws_sqs_queue" "fifo_sqs_queue_withUsage" {
  name       = "my.fifo"
  fifo_queue = true
}
