terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.21.0"
    }
  }
}


resource "aws_sqs_queue" "sqs_withTags_mymod_aliased" {
  name = "sqs_withTags_mymod_aliased"

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
