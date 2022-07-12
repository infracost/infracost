provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "we" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_resource_group" "ne" {
  name     = "example-resources"
  location = "North Europe"
}

resource "azurerm_resource_group" "jw" {
  name     = "example-resources"
  location = "Japan West"
}

resource "azurerm_virtual_network" "we1" {
  name                = "wenetwork1"
  resource_group_name = azurerm_resource_group.we.name
  address_space       = ["10.0.1.0/24"]
  location            = azurerm_resource_group.we.location
}

resource "azurerm_virtual_network" "we2" {
  name                = "wenetwork2"
  resource_group_name = azurerm_resource_group.we.name
  address_space       = ["10.0.2.0/24"]
  location            = azurerm_resource_group.we.location
}

resource "azurerm_virtual_network" "ne" {
  name                = "nenetwork"
  resource_group_name = azurerm_resource_group.ne.name
  address_space       = ["10.0.3.0/24"]
  location            = azurerm_resource_group.ne.location
}

resource "azurerm_virtual_network" "jw" {
  name                = "jwnetwork1"
  resource_group_name = azurerm_resource_group.jw.name
  address_space       = ["10.0.4.0/24"]
  location            = azurerm_resource_group.jw.location
}

resource "azurerm_virtual_network_peering" "intra_region" {
  name                      = "we1towe2"
  resource_group_name       = azurerm_resource_group.we.name
  virtual_network_name      = azurerm_virtual_network.we1.name
  remote_virtual_network_id = azurerm_virtual_network.we2.id
}

resource "azurerm_virtual_network_peering" "intra_zonal" {
  name                      = "wetone"
  resource_group_name       = azurerm_resource_group.we.name
  virtual_network_name      = azurerm_virtual_network.we1.name
  remote_virtual_network_id = azurerm_virtual_network.ne.id
}

resource "azurerm_virtual_network_peering" "inter_zonal" {
  name                      = "wetojw"
  resource_group_name       = azurerm_resource_group.we.name
  virtual_network_name      = azurerm_virtual_network.we1.name
  remote_virtual_network_id = azurerm_virtual_network.jw.id
}

resource "azurerm_virtual_network_peering" "intra_region_with_usage" {
  name                      = "wetowe2"
  resource_group_name       = azurerm_resource_group.we.name
  virtual_network_name      = azurerm_virtual_network.we1.name
  remote_virtual_network_id = azurerm_virtual_network.we2.id
}

resource "azurerm_virtual_network_peering" "intra_zonal_with_usage" {
  name                      = "wetone"
  resource_group_name       = azurerm_resource_group.we.name
  virtual_network_name      = azurerm_virtual_network.we1.name
  remote_virtual_network_id = azurerm_virtual_network.ne.id
}

resource "azurerm_virtual_network_peering" "inter_zonal_with_usage" {
  name                      = "wetojw"
  resource_group_name       = azurerm_resource_group.we.name
  virtual_network_name      = azurerm_virtual_network.we1.name
  remote_virtual_network_id = azurerm_virtual_network.jw.id
}

