provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

locals {
  json_file = jsondecode(file("test.json"))
}

module "defaults" {
  source  = "Invicton-Labs/deepmerge/null"
  version = "0.1.5"
  maps    = [local.json_file]
}

resource "aws_redshift_cluster" "test_from_json_config" {
  count              = module.defaults.merged.test == null ? 0 : 1
  cluster_identifier = "test-infracost-json-config"
  node_type          = "dc2.large"
}
