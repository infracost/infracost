terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.21.0"
    }
  }
}


resource "aws_sqs_queue" "sqs_withTag_mysubmodule" {
  name = "sqs_withTags_mysubmodule"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}
