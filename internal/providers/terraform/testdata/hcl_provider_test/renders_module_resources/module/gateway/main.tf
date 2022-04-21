resource "aws_ec2_transit_gateway" "example" {}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

output "aws_ec2_transit_gateway_id" {
  value = aws_ec2_transit_gateway.example.id
}

output "aws_customer_gateway_id" {
  value = aws_customer_gateway.example.id
}

output "aws_customer_gateway_type" {
  value = aws_customer_gateway.example.type
}
