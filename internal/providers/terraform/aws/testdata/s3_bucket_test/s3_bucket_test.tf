provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_s3_bucket" "bucket1" {
  bucket = "bucket1"

  lifecycle {
    ignore_changes = [
      lifecycle_rule
    ]
  }
}

resource "aws_s3_bucket" "bucket_withUsage" {
  bucket = "bucket_withUsage"

  lifecycle {
    ignore_changes = [
      lifecycle_rule
    ]
  }
}
