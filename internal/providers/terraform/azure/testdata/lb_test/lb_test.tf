provider "azurerm" {
  features {}
  skip_provider_registration = true
}
resource "azurerm_resource_group" "example" {
  name     = "LoadBalancerRG"
  location = "West Europe"
}
resource "azurerm_lb" "basic" {
  name                = "TestLoadBalancer"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_lb" "standard" {
  name                = "TestLoadBalancer"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"
}

resource "azurerm_lb" "withoutUsage" {
  name                = "TestLoadBalancer"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"
}
