provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "merged" {
  source = "./modules/deep_merge"
}

module "merged2" {
  source = "./modules/deep_merge"
}


module "context" {
  source = "./modules/label"
}

output "output" {
  value = {
    "context_name" = module.context.context.name
    "deep_merge_1" = module.merged.merged.key1-3
    "deep_merge_2" = module.merged2.merged.key1-3
  }
}
