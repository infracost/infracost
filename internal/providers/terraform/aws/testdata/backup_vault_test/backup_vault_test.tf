provider "aws" {
  region                      = "us-west-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_backup_vault" "usage" {
  name = "aws_backup_vault"
}

resource "aws_backup_vault" "non_usage" {
  name = "aws_backup_vault"
}
