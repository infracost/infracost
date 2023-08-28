provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for KinesisStream below

resource "aws_kinesis_stream" "test_stream" {
  name             = "terraform-kinesis-test"
  stream_mode_details {
    stream_mode = "ON_DEMAND"
  }
  tags = {
    Environment = "test"
  }
}