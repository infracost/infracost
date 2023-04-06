provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "exampleRG1"
  location = "eastus"
}

resource "azurerm_virtual_network" "example" {
  name                = "example-vnet1"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  address_space       = ["10.0.0.0/22"]
}

resource "azurerm_subnet" "isesubnet1" {
  name                 = "isesubnet1"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
  address_prefixes     = ["10.0.1.0/26"]

  delegation {
    name = "integrationServiceEnvironments"
    service_delegation {
      name = "Microsoft.Logic/integrationServiceEnvironments"
    }
  }
}

resource "azurerm_subnet" "isesubnet2" {
  name                 = "isesubnet2"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
  address_prefixes     = ["10.0.1.64/26"]
}

resource "azurerm_subnet" "isesubnet3" {
  name                 = "isesubnet3"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
  address_prefixes     = ["10.0.1.128/26"]
}

resource "azurerm_subnet" "isesubnet4" {
  name                 = "isesubnet4"
  resource_group_name  = azurerm_resource_group.example.name
  virtual_network_name = azurerm_virtual_network.example.name
  address_prefixes     = ["10.0.1.192/26"]
}

resource "azurerm_integration_service_environment" "example" {

  name                 = "example-ise"
  location             = azurerm_resource_group.example.location
  resource_group_name  = azurerm_resource_group.example.name
  sku_name             = "Premium_3"
  access_endpoint_type = "Internal"
  virtual_network_subnet_ids = [
    azurerm_subnet.isesubnet1.id,
    azurerm_subnet.isesubnet2.id,
    azurerm_subnet.isesubnet3.id,
    azurerm_subnet.isesubnet4.id
  ]
  tags = {
    environment = "development"
  }
}
resource "azurerm_integration_service_environment" "example1" {

  name                 = "example-ise"
  location             = azurerm_resource_group.example.location
  resource_group_name  = azurerm_resource_group.example.name
  sku_name             = "Premium_0"
  access_endpoint_type = "Internal"
  virtual_network_subnet_ids = [
    azurerm_subnet.isesubnet1.id,
    azurerm_subnet.isesubnet2.id,
    azurerm_subnet.isesubnet3.id,
    azurerm_subnet.isesubnet4.id
  ]
  tags = {
    environment = "development"
  }
}
resource "azurerm_integration_service_environment" "example2" {

  name                 = "example-ise"
  location             = azurerm_resource_group.example.location
  resource_group_name  = azurerm_resource_group.example.name
  sku_name             = "Developer_0"
  access_endpoint_type = "Internal"
  virtual_network_subnet_ids = [
    azurerm_subnet.isesubnet1.id,
    azurerm_subnet.isesubnet2.id,
    azurerm_subnet.isesubnet3.id,
    azurerm_subnet.isesubnet4.id
  ]
  tags = {
    environment = "development"
  }
}
