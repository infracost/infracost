provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_efs_file_system" "standard" {
  lifecycle_policy {
    transition_to_ia = "AFTER_7_DAYS"
  }
}

resource "aws_efs_file_system" "oneZone" {
  lifecycle_policy {
    transition_to_ia = "AFTER_7_DAYS"
  }
  availability_zone_name = "One Zone"
}

resource "aws_efs_file_system" "provisioned" {
  provisioned_throughput_in_mibps = 100
  throughput_mode                 = "provisioned"
}

resource "aws_efs_file_system" "no_usage" {
  availability_zone_name = "One Zone"
}

resource "aws_efs_file_system" "no_usage_with_lifecycle_policy" {
  lifecycle_policy {
    transition_to_ia = "AFTER_7_DAYS"
  }
  availability_zone_name = "One Zone"
}
