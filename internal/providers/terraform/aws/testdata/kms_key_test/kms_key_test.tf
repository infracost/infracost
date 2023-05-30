provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_kms_key" "kms" {}

resource "aws_kms_key" "rsa2048" {
  customer_master_key_spec = "RSA_2048"
}

resource "aws_kms_key" "rsa3072" {
  customer_master_key_spec = "RSA_3072"
}
