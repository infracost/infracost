module "mymod" {
  source     = "./module"
  aws_region = "us-east-1"
  tags       = local.unknown_var
}
