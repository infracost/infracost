provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_powerbi_embedded" "a1" {
  name                = "examplepowerbi"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku_name            = "A1"
  administrators      = ["azsdktest@microsoft.com"]
}

resource "azurerm_powerbi_embedded" "a5" {
  name                = "examplepowerbi"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku_name            = "A5"
  administrators      = ["azsdktest@microsoft.com"]
}
