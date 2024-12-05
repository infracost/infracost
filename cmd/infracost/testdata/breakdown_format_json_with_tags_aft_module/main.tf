provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"

  default_tags {
    tags = {
      DefaultNotOverride = "defaultnotoverride-aws"
      DefaultOverride    = "defaultoverride-aws"
    }
  }
}

module "aft" {
  source = "github.com/aws-ia/terraform-aws-control_tower_account_factory?ref=1.13.2"
  # Required Variables
  ct_management_account_id    = "1234"
  log_archive_account_id      = "1234"
  audit_account_id            = "1234"
  aft_management_account_id   = "1234"
  ct_home_region              = "eu-central-1"
  tf_backend_secondary_region = "eu-west-1"

  # Terraform variables
  terraform_distribution = "oss"
  terraform_version      = "1.4.6"

  # VCS variables
  vcs_provider                                  = "github"
  account_request_repo_name                     = "myrepo/aft-account-requests"
  account_provisioning_customizations_repo_name = "myrepo/aft-account-provisioning-customizations"
  global_customizations_repo_name               = "myrepo/aft-global-customizations"
  account_customizations_repo_name              = "myrepo/aft-account-customizations"

  # AFT Feature flags
  aft_feature_cloudtrail_data_events      = false // all new accounts provisioned via AFT will have CloudTrail data events disabled
  aft_feature_enterprise_support          = true  // all new accounts provisioned via AFT will be enrolled in enterprise support
  aft_feature_delete_default_vpcs_enabled = true  // all new accounts provisioned via AFT will not contain the default VPC
}
