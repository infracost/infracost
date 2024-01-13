provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride-aws.env"
      DefaultOverride    = "defaultoverride-aws.env"
    }
  }

  alias = "env"
}

provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride-aws"
      DefaultOverride    = "defaultoverride-aws"
    }
  }
}


module "mymod_aliased" {
  source = "./mymod_aliased"

  providers = {
    aws = aws.env
  }
}

module "mymod_implicit" {
  source = "./mymod_implicit"
}

module "mymod_own" {
  source = "./mymod_own"
}
