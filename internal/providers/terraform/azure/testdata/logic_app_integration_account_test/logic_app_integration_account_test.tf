provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

locals {
  skus = ["Free", "Basic", "Standard"]
}

resource "azurerm_logic_app_integration_account" "example" {
  for_each = { for sku in local.skus : sku => sku }

  name                = "account-${each.value}"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_name            = each.value
}
