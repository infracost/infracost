provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

// Scenario 1: No usage, standard license (no association)
resource "aws_grafana_workspace" "no_license" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = "no-license"
}

// Scenario 3: No usage, enterprise license
resource "aws_grafana_workspace" "with_enterprise_license" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = "with-enterprise-license"
}

resource "aws_grafana_license_association" "enterprise" {
  license_type = "ENTERPRISE"
  workspace_id = aws_grafana_workspace.with_enterprise_license.id
}

// For usage file scenarios
// Scenario 2: With usage, standard license
resource "aws_grafana_workspace" "with_usage" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = "with-usage"
}

// Scenario 4: With usage, enterprise license
resource "aws_grafana_workspace" "with_enterprise_license_and_usage" {
  account_access_type      = "CURRENT_ACCOUNT"
  authentication_providers = ["SAML"]
  permission_type          = "SERVICE_MANAGED"
  name                     = "with-enterprise-license-and-usage"
}

resource "aws_grafana_license_association" "enterprise_usage" {
  license_type = "ENTERPRISE"
  workspace_id = aws_grafana_workspace.with_enterprise_license_and_usage.id
}
