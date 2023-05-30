provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_ec2_client_vpn_endpoint" "endpoint" {
  description            = "terraform-clientvpn-example"
  server_certificate_arn = "arn:aws:acm:us-east-1:123456789123:certificate/a13e05dc-c58d-43f8-8a9b-c456f67891c2"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "arn:aws:acm:us-east-1:123456789123:certificate/a13e05dc-c58d-43f8-8a9b-c456f67891c2"
  }

  connection_log_options {
    enabled               = true
    cloudwatch_log_group  = "cloudwatch-log-group"
    cloudwatch_log_stream = "cloudwatch-log-group-stream"
  }
}
