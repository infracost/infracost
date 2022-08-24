locals {
  aws_region = "us-east-1"
}

iam_role = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"

# Generate an AWS provider block
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
provider "aws" {
  region                      = "${local.aws_region}"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}
EOF
}
