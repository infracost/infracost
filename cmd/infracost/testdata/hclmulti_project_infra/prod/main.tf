provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "base" {
  source = "../modules/base"

  env_name = "prod"
}

module "back_api" {
  source                      = "../modules/back_api"
  env_name                    = "prod"
  eks_assume_role_policy      = ""
  api_keys_table_arn          = ""
  database_instance_type      = "db.t3.small"
  database_subnet_group_name  = module.base.database_subnet_group_name
  database_security_group_ids = []
}

module "front" {
  source = "../modules/front"

  env_name                    = "prod"
  vpc_id                      = module.base.vpc_id
  database_subnet_group_name  = module.base.database_subnet_group_name
  database_security_group_ids = []
  eks_assume_role_policy      = ""
  root_domain                 = "my.com"
  api_keys_table_arn          = ""
  wildcard_cert_arn           = ""
  front_web_domain            = "front.my.com"
}

output "front_db_url" {
  value     = module.front.front_db_url
  sensitive = true
}

output "front_db_password" {
  value     = module.front.front_db_password
  sensitive = true
}

output "front_web_bucket" {
  value = module.front.front_web_bucket
}

output "back_api_role_arn" {
  value = module.back_api.role_arn
}

output "back_api_db_data_bucket_name" {
  value = module.back_api.db_data_bucket_name
}

output "back_api_db_address" {
  value = module.back_api.back_api_db_address
}

output "back_api_db_password" {
  value     = module.back_api.back_api_db_password
  sensitive = true
}
