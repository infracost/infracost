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
}

resource "aws_s3_bucket" "bucket2" {
  bucket = "bucket2"
}

resource "aws_s3_bucket_inventory" "inventory" {
  bucket                   = aws_s3_bucket.bucket1.bucket
  name                     = "inventory"
  included_object_versions = "All"

  schedule {
    frequency = "Daily"
  }

  destination {
    bucket {
      format     = "CSV"
      bucket_arn = aws_s3_bucket.bucket2.arn
    }
  }
}

resource "aws_s3_bucket" "bucket1_withUsage" {
  bucket = "bucket1_withUsage"
}

resource "aws_s3_bucket" "bucket2_withUsage" {
  bucket = "bucket2_withUsage"
}

resource "aws_s3_bucket_inventory" "inventory_withUsage" {
  bucket                   = aws_s3_bucket.bucket1_withUsage.bucket
  name                     = "inventory_withUsage"
  included_object_versions = "All"

  schedule {
    frequency = "Daily"
  }

  destination {
    bucket {
      format     = "CSV"
      bucket_arn = aws_s3_bucket.bucket2_withUsage.arn
    }
  }
}
