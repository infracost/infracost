terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.21.0"
    }
  }
}

resource "aws_sqs_queue" "sqs_withTag_mysubsub_aliased" {
  name = "sqs_withTag_mysubsub_aliased"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}
