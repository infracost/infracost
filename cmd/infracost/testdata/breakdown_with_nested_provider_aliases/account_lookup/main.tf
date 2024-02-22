module "region_data" {
  source = "../region_lookup"
}

output "data" {
  value = {
    "role"        = "arn:aws:iam::12345:role/role"
    "region_name" = module.region_data.full_name
  }
}
