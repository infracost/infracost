provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "rg" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_container_registry" "my_registry" {
  name                = "containerRegistry1"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Premium"
  admin_enabled       = false

  georeplications {
    location = "East US"
  }
}

resource "azurerm_container_registry" "twoLocations" {
  name                = "containerRegistry1"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Premium"
  admin_enabled       = false

  georeplications {
    location = "East US"
  }

  georeplications {
    location = "North Europe"
  }
}

resource "azurerm_container_registry" "withoutLocations" {
  name                = "containerRegistry1"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Premium"
  admin_enabled       = false
}

resource "azurerm_container_registry" "basic" {
  name                = "containerRegistry1"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
  admin_enabled       = false
}

resource "azurerm_container_registry" "withoutUsage" {
  name                = "containerRegistry1"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
  admin_enabled       = false
}

resource "azurerm_container_registry" "overStorage" {
  name                = "containerRegistry1"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
  admin_enabled       = false
}
