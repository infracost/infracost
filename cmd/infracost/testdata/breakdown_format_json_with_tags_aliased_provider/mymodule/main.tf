terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.21.0"
    }
  }
}


resource "aws_sqs_queue" "sqs_withTags_mymodule" {
  name = "sqs_withTags_mymodule"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}

module "mysub-explict-provider" {
  source = "../mysubmodule"

  providers = {
    aws = aws
  }
}

module "mysub-implict-provider" {
  source = "../mysubmodule"
}
