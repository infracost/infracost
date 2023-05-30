provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}
resource "aws_ec2_transit_gateway_peering_attachment" "peering" {
  peer_account_id         = "123456789111"
  peer_region             = "eu-west-1"
  peer_transit_gateway_id = "tgw-654321"
  transit_gateway_id      = "tgw-123456"
}
