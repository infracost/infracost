terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.21.0"
    }
  }
}

resource "aws_sqs_queue" "sqs_withTag_mysub_implicit" {
  name = "sqs_withTag_mysub_implicit"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}

module "mysubsub_implict" {
  source = "../../mysubsub_implicit"
}

module "mysubsub_own" {
  source = "../../mysubsub_own"
}
