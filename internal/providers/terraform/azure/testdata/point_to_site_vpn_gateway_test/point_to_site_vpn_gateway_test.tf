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

resource "azurerm_vpn_server_configuration" "server_config" {
  name                = "example-config"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  vpn_authentication_types = [
  "Certificate"]
}


resource "azurerm_point_to_site_vpn_gateway" "point_to_site_vpn" {
  name                        = "example-vpn"
  location                    = azurerm_resource_group.resource_group.location
  resource_group_name         = azurerm_resource_group.resource_group.name
  virtual_hub_id              = azurerm_virtual_hub.virtual_hub.id
  scale_unit                  = 5
  vpn_server_configuration_id = azurerm_vpn_server_configuration.server_config.id
  connection_configuration {
    name = "example-gateway-config"

    vpn_client_address_pool {
      address_prefixes = [
        "10.0.2.0/24"
      ]
    }
  }
}

resource "azurerm_point_to_site_vpn_gateway" "point_to_site_vpn_with_usage" {
  name                        = "example-vpn"
  location                    = azurerm_resource_group.resource_group.location
  resource_group_name         = azurerm_resource_group.resource_group.name
  virtual_hub_id              = azurerm_virtual_hub.virtual_hub.id
  scale_unit                  = 5
  vpn_server_configuration_id = azurerm_vpn_server_configuration.server_config.id
  connection_configuration {
    name = "example-gateway-config"

    vpn_client_address_pool {
      address_prefixes = [
        "10.0.2.0/24"
      ]
    }
  }
}
