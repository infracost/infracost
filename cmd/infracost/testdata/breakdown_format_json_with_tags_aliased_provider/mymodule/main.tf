terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.21.0"
    }
  }
}


resource "aws_sqs_queue" "sqs_withTags" {
  name = "sqs_withTags"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}