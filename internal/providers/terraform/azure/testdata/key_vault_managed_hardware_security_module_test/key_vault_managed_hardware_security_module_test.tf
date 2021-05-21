provider "azurerm" {
  features {}
  skip_provider_registration = true
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

data "azurerm_client_config" "current" {
}

resource "azurerm_key_vault_managed_hardware_security_module" "my_module" {
  name                       = "exampleKVHsm"
  resource_group_name        = azurerm_resource_group.example.name
  location                   = "eastus"
  sku_name                   = "Standard_B1"
  purge_protection_enabled   = false
  soft_delete_retention_days = 90
  tenant_id                  = "00000000-0000-0000-0000-000000000000"
  admin_object_ids           = [data.azurerm_client_config.current.object_id]
}
