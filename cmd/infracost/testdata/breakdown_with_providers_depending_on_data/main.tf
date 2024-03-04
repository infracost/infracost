provider "aws" {
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
  region                      = module.mod_us2.region_us2
}

provider "aws" {
  alias                       = "eu1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
  region                      = module.mod_eu1.region_eu1

}

module "mod_us2" {
  source = "./mod"
}

module "mod_eu1" {
  source = "./mod"

  providers = {
    aws = aws.eu1
  }
}
