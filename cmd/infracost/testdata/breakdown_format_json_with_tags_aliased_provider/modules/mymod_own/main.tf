provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride-own-provider"
      DefaultOverride    = "defaultoverride-own-provider"
    }
  }
}

resource "aws_sqs_queue" "sqs_withTags_mymod_own" {
  name = "sqs_withTags_mymod_own"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}

module "mysub_aliased" {
  source = "./mysub_aliased"

  providers = {
    aws = aws
  }
}

module "mysub_implicit" {
  source = "./mysub_implicit"
}
