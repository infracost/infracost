locals {
  names_map = [{ name = "name-a", name = "name-b" }]
}

module "gateway_proxy" {
  for_each = local.names_map

  source = "./../gateway"

  tags = {
    Name = each.value.name
  }
}
