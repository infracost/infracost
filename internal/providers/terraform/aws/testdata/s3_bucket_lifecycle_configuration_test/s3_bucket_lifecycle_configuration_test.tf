provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_s3_bucket" "bucket" {
  bucket = "test-bucket"

  lifecycle {
    ignore_changes = [
      lifecycle_rule
    ]
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "bucket_lifecycle_config" {
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    id     = "rule1"
    status = "Enabled"

    filter {
      tag {
        key   = "key"
        value = "value"
      }
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

  rule {
    id     = "rule2"
    status = "Enabled"

    filter {
      tag {
        key   = "key"
        value = "value"
      }
    }

    noncurrent_version_transition {
      storage_class = "INTELLIGENT_TIERING"
    }
    noncurrent_version_transition {
      storage_class = "ONEZONE_IA"
    }
    noncurrent_version_transition {
      storage_class = "STANDARD_IA"
    }
    noncurrent_version_transition {
      storage_class = "GLACIER"
    }
    noncurrent_version_transition {
      storage_class = "DEEP_ARCHIVE"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "bucket_lifecycle_config_noncurrent_version_transitions" {
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    id     = "rule1"
    status = "Enabled"

    filter {
      and {
        tags = {
          tag1 = "value1"
          tag2 = "value2"
        }
      }
    }

    noncurrent_version_transition {
      storage_class = "INTELLIGENT_TIERING"
    }
    noncurrent_version_transition {
      storage_class = "ONEZONE_IA"
    }
    noncurrent_version_transition {
      storage_class = "STANDARD_IA"
    }
    noncurrent_version_transition {
      storage_class = "GLACIER"
    }
    noncurrent_version_transition {
      storage_class = "DEEP_ARCHIVE"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "bucket_lifecycle_config_with_usage" {
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    id     = "rule1"
    status = "Enabled"

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

    filter {
      tag {
        key   = "key"
        value = "value"
      }
    }
  }

  rule {
    id     = "rule2"
    status = "Disabled"

    transition {
      storage_class = "DEEP_ARCHIVE"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "bucket_lifecycle_config_without_tags" {
  bucket = aws_s3_bucket.bucket.bucket

  rule {
    id     = "rule1"
    status = "Enabled"
  }
}
