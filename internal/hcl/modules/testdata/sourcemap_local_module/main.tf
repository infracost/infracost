provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "local-module" {
  source        = "../sourcemap_local_module/./modules/local-module"
  instance_type = "m5.4xlarge"
}
