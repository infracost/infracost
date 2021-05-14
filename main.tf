provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region = "us-central1"
}

provider "aws" {
  region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

provider "azurerm" {
  skip_provider_registration = true
  features {}

  subscription_id             = "84171726-c002-4fc7-924a-b6ff82c59677"
  client_id                   = "eec7409b-6cbd-420d-b3ba-bc0ef14c4bbf"
  client_secret               = "OZ~wK2ZpptK2O1Z61_qD2ycu~1KQ4WDzvL"
  tenant_id                   = "b0b4f93e-59ae-44ff-b39a-e51e7e2b16e9"
}

resource "azurerm_resource_group" "rg" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_container_registry" "my_registry" {
  name                     = "containerRegistry1"
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = azurerm_resource_group.rg.location
  sku                      = "Standard"
  admin_enabled            = false
  georeplication_locations = ["East US"]
}

