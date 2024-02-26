terraform {
  source = "./modules"
}

locals {
  aws_region = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
}

# Generate an AWS provider block
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
provider "aws" {
  # This missing thing shouldn't make this generate fail: ${local.missing.thing}
  region                      = "${local.aws_region}"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride"
      DefaultOverride    = "defaultoverride"
    }
  }
}
EOF
}
