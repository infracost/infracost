terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.21.0"
    }
  }
}

resource "aws_sqs_queue" "sqs_withTags_mysub_aliased" {
  name = "sqs_withTags_mysub_aliased"

  tags = {
    DefaultOverride = "sqs-def"
    ResourceTag     = "sqs-hi"
  }
}

module "mysubsub_aliased" {
  source = "../../mysubsub_aliased"

  providers = {
    aws = aws
  }
}

module "mysubsub_implict" {
  source = "../../mysubsub_implicit"
}

module "mysubsub_own" {
  source = "../../mysubsub_own"
}
