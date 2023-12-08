provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# Add example resources for CloudHSMv2HSM below

resource "aws_cloudhsm_v2_hsm" "cloudhsm_v2_hsm" {
  subnet_id  = "subnet-123456"
  cluster_id = "cluster-id"
}

resource "aws_cloudhsm_v2_hsm" "cloudhsm_v2_hsm_with_usage" {
  subnet_id  = "subnet-123456"
  cluster_id = "cluster-id"
}
