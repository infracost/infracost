provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_iothub_device_update_account" "example_account" {
  name                = "example-update-account"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_iothub" "iothub_single" {
  name                = "example-iothub"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  sku {
    name     = "S1"
    capacity = 1
  }

  tags = {
    purpose = "testing"
  }
}

resource "azurerm_iothub_device_update_instance" "standard_instance" {
  name                     = "example-update-instance"
  device_update_account_id = azurerm_iothub_device_update_account.example_account.id
  iothub_id                = azurerm_iothub.iothub_single.id
}
