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
  address_prefix      = "10.0.1.0/24"
}

resource "azurerm_express_route_gateway" "express_route" {
  name                = "express-route"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  virtual_hub_id      = azurerm_virtual_hub.virtual_hub.id
  scale_units         = 1
}

resource "azurerm_express_route_port" "route_port" {
  name                = "example-erp"
  resource_group_name = azurerm_resource_group.resource_group.name
  location            = azurerm_resource_group.resource_group.location
  peering_location    = "Equinix-Seattle-SE2"
  bandwidth_in_gbps   = 10
  encapsulation       = "Dot1Q"
}

resource "azurerm_express_route_circuit" "route_circuit" {
  name                  = "example-erc"
  resource_group_name   = azurerm_resource_group.resource_group.name
  location              = azurerm_resource_group.resource_group.location
  express_route_port_id = azurerm_express_route_port.route_port.id
  bandwidth_in_gbps     = 5

  sku {
    tier   = "Standard"
    family = "MeteredData"
  }
}

resource "azurerm_express_route_circuit_peering" "circuit_peering" {
  peering_type                  = "AzurePrivatePeering"
  express_route_circuit_name    = azurerm_express_route_circuit.route_circuit.name
  resource_group_name           = azurerm_resource_group.resource_group.name
  shared_key                    = "ItsASecret"
  peer_asn                      = 100
  primary_peer_address_prefix   = "192.168.1.0/30"
  secondary_peer_address_prefix = "192.168.2.0/30"
  vlan_id                       = 100
}

resource "azurerm_express_route_connection" "express_route_conn" {
  name                             = "express-route-conn"
  express_route_circuit_peering_id = azurerm_express_route_circuit_peering.circuit_peering.id
  express_route_gateway_id         = azurerm_express_route_gateway.express_route.id
}
