provider "azurerm" {
  skip_provider_registration = true
  features {}
}


resource "azurerm_resource_group" "resource_group" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_virtual_wan" "virtual_wan" {
  name                = "example-virtualwan"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
}

resource "azurerm_virtual_hub" "virtual_hub" {
  name                = "example-virtualhub"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  virtual_wan_id      = azurerm_virtual_wan.virtual_wan.id
  sku                 = "Standard"
}

resource "azurerm_virtual_hub" "free_virtual_hub_by_default" {
  name                = "example-virtualhub"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  virtual_wan_id      = azurerm_virtual_wan.virtual_wan.id
}

resource "azurerm_virtual_hub" "free_virtual_hub_by_specification" {
  name                = "example-virtualhub"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  virtual_wan_id      = azurerm_virtual_wan.virtual_wan.id
  sku                 = "Basic"
}

resource "azurerm_virtual_hub" "virtual_hub_with_usage" {
  name                = "example-virtualhub"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  virtual_wan_id      = azurerm_virtual_wan.virtual_wan.id
  sku                 = "Standard"
}
