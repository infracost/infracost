# https://github.com/localstack/localstack could also be used to speed-up dev/test
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
    }
    # TODO Enable once we have released terraform-provider-infracost
    # infracost = {
    #   source = "infracost.io/infracost/infracost"
    # }
  }
}

provider "aws" {
  region                      = "us-east-1"
  s3_force_path_style         = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

# TODO Enable once we have released terraform-provider-infracost
# provider "infracost" {}

data "aws_region" "current" {}

variable "availability_zone_names" {
  type    = list(string)
  default = ["us-east-1a", "us-east-1b"]
}

variable "aws_subnet_ids" {
  type    = list(string)
  default = ["fake1", "fake2"]
}

variable "aws_ami_id" {
  type    = string
  default = "fake1"
}
