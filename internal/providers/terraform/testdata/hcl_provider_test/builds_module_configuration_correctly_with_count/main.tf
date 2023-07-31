provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      Version = "1.0.0"
    }
  }
}

module "gateway_test" {
  source = "./module/gateway_proxy"
}
