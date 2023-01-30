provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

module "registry-submodule" {
  source  = "terraform-aws-modules/route53/aws//modules/zones"
  version = "2.5.0"

  zones = {
    "example.com" = {
      comment = "example.com"
    }
  }
}

module "registry-submodule-records" {
  source  = "terraform-aws-modules/route53/aws//modules/records"
  version = "2.5.0"
}

module "git-submodule" {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-route53.git//modules/zones"

  zones = {
    "example.com" = {
      comment = "example.com"
    }
  }
}
