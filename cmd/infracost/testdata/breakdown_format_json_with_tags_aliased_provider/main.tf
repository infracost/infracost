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


module "mymod_provider_alias" {
  source = "./mymodule"

  providers = {
    aws = aws.env
  }
}

module "mymode_implicit_providers" {
  source = "./mymodule"
}

