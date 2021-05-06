provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_vpn_connection" "vpn_connection" {
  customer_gateway_id = "dummy-customer-gateway-id"
  type = "ipsec.1"
}

resource "aws_vpn_connection" "transit" {
  customer_gateway_id = "dummy-customer-gateway-id"
  type = "ipsec.1"
  transit_gateway_id = "dummy-transit-gateway-id"
}

resource "aws_vpn_connection" "vpn_connection_withUsage" {
  customer_gateway_id = "dummy-customer-gateway-id2"
  type = "ipsec.1"
}

resource "aws_vpn_connection" "transit_withUsage" {
  customer_gateway_id = "dummy-customer-gateway-id2"
  type = "ipsec.1"
  transit_gateway_id = "dummy-transit-gateway-id2"
}