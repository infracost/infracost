provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  alias = "own_aliased"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride-own-aliased-provider"
      DefaultOverride    = "defaultoverride-own-aliased-provider"
    }
  }
}

resource "aws_sqs_queue" "sqs_withTags_mymod_own_aliased" {
  name     = "sqs_withTags_mymod_own_aliased"
  provider = aws.own_aliased

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}

module "mysub_aliased" {
  source = "./mysub_aliased"

  providers = {
    aws.own_aliased = aws.own_aliased
  }
}

module "mysub_implicit" {
  source = "./mysub_implicit"
}
