provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_cloudtrail" "cloudtrail_with_defaults" {
  name           = "cloudtrail-with-defaults"
  s3_bucket_name = "bucket-1234"
}

resource "aws_cloudtrail" "cloudtrail_without_management_events" {
  name                          = "cloudtrail-with-defaults"
  s3_bucket_name                = "bucket-1234"
  include_global_service_events = false
}

resource "aws_cloudtrail" "cloudtrail_with_insight_selector" {
  name           = "cloudtrail-with-defaults"
  s3_bucket_name = "bucket-1234"
  insight_selector {
    insight_type = "ApiCallRateInsight"
  }
}

resource "aws_cloudtrail" "cloudtrail_with_usage" {
  name           = "cloudtrail-with-defaults"
  s3_bucket_name = "bucket-1234"
  insight_selector {
    insight_type = "ApiCallRateInsight"
  }
}
