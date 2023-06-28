provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_vpn_connection" "vpn_connection" {
  customer_gateway_id = "dummy-customer-gateway-id"
  type                = "ipsec.1"
  vpn_gateway_id      = "vpn-gateway-id"
}

resource "aws_vpn_connection" "transit" {
  customer_gateway_id = "dummy-customer-gateway-id"
  type                = "ipsec.1"
  transit_gateway_id  = "dummy-transit-gateway-id"
}

resource "aws_vpn_connection" "vpn_connection_withUsage" {
  customer_gateway_id = "dummy-customer-gateway-id2"
  type                = "ipsec.1"
  vpn_gateway_id      = "vpn-gateway-id"
}

resource "aws_vpn_connection" "transit_withUsage" {
  customer_gateway_id = "dummy-customer-gateway-id2"
  type                = "ipsec.1"
  transit_gateway_id  = "dummy-transit-gateway-id2"
}
