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
}

resource "azurerm_vpn_gateway" "default_vpn" {
  name                = "example-vpn"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  virtual_hub_id      = azurerm_virtual_hub.virtual_hub.id
}

resource "azurerm_vpn_gateway" "vpn_with_scale_units" {
  name                = "example-vpn-scale"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  virtual_hub_id      = azurerm_virtual_hub.virtual_hub.id
  scale_unit          = 3
}
