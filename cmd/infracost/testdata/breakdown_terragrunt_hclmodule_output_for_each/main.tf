provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "mod1" {
  source = "./mod1"
}

module "mod2" {
  for_each = { for k, v in module.mod1.instances : k => v }

  source = "./mod2"

  instance_type = each.value
}
