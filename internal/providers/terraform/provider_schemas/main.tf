# This file is used with the "terraform provider schema" commands in "make tagschema" to generate
# the list of terraform resources that support tags.

provider "aws" {
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

provider "azurerm" {
  skip_provider_registration = true
  features {}
}

provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  project     = "my-project"
}
