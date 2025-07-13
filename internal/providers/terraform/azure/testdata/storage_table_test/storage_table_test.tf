provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "test-resources"
  location = "West Europe"
}

resource "azurerm_storage_account" "test" {
  name                     = "infracosttestsa"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  account_kind             = "StorageV2"
}

resource "azurerm_storage_table" "test" {
  name                 = "testtable"
  storage_account_name = azurerm_storage_account.test.name
} 