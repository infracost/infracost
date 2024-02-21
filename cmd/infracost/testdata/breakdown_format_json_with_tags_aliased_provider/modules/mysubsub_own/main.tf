provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride-subsub-own-provider"
      DefaultOverride    = "defaultoverride-subsub-own-provider"
    }
  }
}

resource "aws_sqs_queue" "sqs_withTag_mysubsub_own" {
  name = "sqs_withTag_mysubsub_own"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}
