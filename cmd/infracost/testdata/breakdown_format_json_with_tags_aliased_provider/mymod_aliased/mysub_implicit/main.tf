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

# can't use the mysubub_aliased module here because we don't have an explicit 'aws'
# provider. terraform init gives this warning:
#   The configuration for module.mymod_aliased.module.mysub_implicit expects to
#   inherit a configuration for provider hashicorp/aws with local name "aws",
#   but module.mymod_aliased doesn't pass a configuration under that name.
#module "mysubsub_aliased" {
#  source = "../../mysubsub_aliased"
#
#  providers = {
#    aws = aws
#  }
#}

module "mysubsub_implict" {
  source = "../../mysubsub_implicit"
}

module "mysubsub_own" {
  source = "../../mysubsub_own"
}
