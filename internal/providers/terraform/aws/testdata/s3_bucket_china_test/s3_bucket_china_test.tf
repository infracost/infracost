provider "aws" {
  region                      = "cn-north-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_s3_bucket" "bucket_china" {
  bucket = "bucket-china"

  lifecycle {
    ignore_changes = [
      lifecycle_rule
    ]
  }
}

resource "aws_s3_bucket" "bucket_china_with_usage" {
  bucket = "bucket_china_with_usage"

  lifecycle {
    ignore_changes = [
      lifecycle_rule
    ]
  }
}

