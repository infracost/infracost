variable "tags" {
  type    = map(string)
  default = {}
}

locals {
  tags = merge(var.tags, {
    "resource_tag" = "tag"
  })
}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"

  tags = local.tags
}
