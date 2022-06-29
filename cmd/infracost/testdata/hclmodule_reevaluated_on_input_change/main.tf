terraform {
  required_version = ">= 0.13"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 2.0"
    }
    template = {
      source  = "hashicorp/template"
      version = ">= 2.0"
    }
  }
}

locals {
  enabled = module.out.enabled
}

module "out" {
  source = "./modules/out"
}

module "test" {
  source  = "./modules/test"
  enabled = local.enabled
}
