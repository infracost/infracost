terraform {
  required_providers {
    aws-v3 = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "aws-v3" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_s3_bucket" "bucket1" {
  provider = aws-v3
  bucket   = "bucket1"

  lifecycle_rule {
    enabled = true
    tags = {
      Key = "value"
    }

    transition {
      storage_class = "INTELLIGENT_TIERING"
    }
    transition {
      storage_class = "ONEZONE_IA"
    }
    transition {
      storage_class = "STANDARD_IA"
    }
    transition {
      storage_class = "GLACIER"
    }
    transition {
      storage_class = "DEEP_ARCHIVE"
    }
  }
}

resource "aws_s3_bucket" "bucket_withUsage" {
  provider = aws-v3
  bucket   = "bucket_withUsage"

  lifecycle_rule {
    enabled = true
    tags = {
      Key = "value"
    }
  }
}
