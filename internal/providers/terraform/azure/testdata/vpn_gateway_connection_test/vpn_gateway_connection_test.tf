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

resource "azurerm_vpn_gateway" "vpn_gateway" {
  name                = "example-vpn"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  virtual_hub_id      = azurerm_virtual_hub.virtual_hub.id
}

resource "azurerm_vpn_site" "vpn_site" {
  name                = "example-vpn-site"
  location            = azurerm_resource_group.resource_group.location
  resource_group_name = azurerm_resource_group.resource_group.name
  virtual_wan_id      = azurerm_virtual_wan.virtual_wan.id
  link {
    name       = "link1"
    ip_address = "10.1.0.0"
  }
  link {
    name       = "link2"
    ip_address = "10.2.0.0"
  }
}

resource "azurerm_vpn_gateway_connection" "vpn_gateway_connection" {
  name               = "example-vpn-gateway-conn"
  vpn_gateway_id     = azurerm_vpn_gateway.vpn_gateway.id
  remote_vpn_site_id = azurerm_vpn_site.vpn_site.id

  vpn_link {
    name             = "link1"
    vpn_site_link_id = azurerm_vpn_site.vpn_site.link[0].id
  }
}
