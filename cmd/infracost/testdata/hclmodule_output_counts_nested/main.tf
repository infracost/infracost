provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "sub" {
  source  = "./modules/sub"
  enabled = module.this.enabled
}

module "this" {
  source  = "./modules/test"
  enabled = false
}
