provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}

resource "azurerm_static_site" "free" {
  name                = "example-static-site-free"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_tier            = "Free"
}

resource "azurerm_static_site" "standard" {
  name                = "example-static-site-standard"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  sku_tier            = "Standard"
}

resource "azurerm_static_site" "default" {
  name                = "example-static-site-default"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
} 