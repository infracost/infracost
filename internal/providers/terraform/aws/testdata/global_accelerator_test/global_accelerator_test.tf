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

# Add example resources for GlobalAccelerator below

resource "aws_globalaccelerator_accelerator" "example" {
  name            = "Example"
  ip_address_type = "IPV4"
  enabled         = true

  #attributes {
    #flow_logs_enabled   = true
    #flow_logs_s3_bucket = "example-bucket"
    #flow_logs_s3_prefix = "flow-logs/"
  #}
}
