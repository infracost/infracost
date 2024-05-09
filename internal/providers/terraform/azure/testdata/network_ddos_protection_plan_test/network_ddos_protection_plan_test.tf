provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}


resource "azurerm_network_ddos_protection_plan" "network_ddos_protection_plan" {
  name                = "exampleresource"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
}

resource "azurerm_network_ddos_protection_plan" "network_ddos_protection_plan_with_usage" {
  name                = "exampleresource"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
}
