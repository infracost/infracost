provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "gateway" {
  source = "./module/gateway"
}

resource "aws_vpn_connection" "example" {
  customer_gateway_id = module.gateway.aws_customer_gateway_id
  transit_gateway_id  = module.gateway.aws_ec2_transit_gateway_id
  type                = module.gateway.aws_customer_gateway_type
}
