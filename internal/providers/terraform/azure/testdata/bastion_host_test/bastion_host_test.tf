provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "deploy" {
  name     = "example-resources"
  location = "australiaeast"
}

resource "azurerm_virtual_network" "deploy" {
  name                = "deploy-vnet"
  location            = azurerm_resource_group.deploy.location
  resource_group_name = azurerm_resource_group.deploy.name
  address_space       = ["10.0.1.0/16"]
}

resource "azurerm_subnet" "deploy" {
  name                 = "deploy-subnet"
  resource_group_name  = azurerm_resource_group.deploy.name
  virtual_network_name = azurerm_virtual_network.deploy.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_public_ip" "deploy" {
  name                = "deploy-ip"
  location            = azurerm_resource_group.deploy.location
  resource_group_name = azurerm_resource_group.deploy.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_bastion_host" "example" {
  name                = "examplebastion"
  location            = azurerm_resource_group.deploy.location
  resource_group_name = azurerm_resource_group.deploy.name

  ip_configuration {
    name                 = "configuration"
    subnet_id            = azurerm_subnet.deploy.id
    public_ip_address_id = azurerm_public_ip.deploy.id
  }
}
