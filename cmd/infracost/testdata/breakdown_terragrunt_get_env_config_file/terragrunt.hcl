locals {
  env = get_env("ENV")
}

terraform {
  source = "./main.tf"
}

inputs = {
  env    = local.env
}
