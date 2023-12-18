provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "git-submodule" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git//modules/zones?ref=v2.5.0"

  zones = {
    "example.com" = {
      comment = "example.com"
    }
  }
}
