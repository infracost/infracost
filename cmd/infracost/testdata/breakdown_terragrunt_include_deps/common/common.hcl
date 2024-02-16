locals {
    foo = "bar"
}

dependency "baz" {
  config_path = "${get_original_terragrunt_dir()}/../baz"
}
