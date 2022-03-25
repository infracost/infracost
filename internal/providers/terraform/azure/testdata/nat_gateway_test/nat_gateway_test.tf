provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "nat-gateway-example-rg"
  location = "West Europe"
}

resource "azurerm_public_ip" "example" {
  name                = "nat-gateway-publicIP"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  allocation_method   = "Static"
  sku                 = "Standard"
  zones               = ["1"]
}

resource "azurerm_public_ip_prefix" "example" {
  name                = "nat-gateway-publicIPPrefix"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  prefix_length       = 30
  zones               = ["1"]
}

resource "azurerm_nat_gateway" "my_gateway" {
  name                    = "nat-Gateway"
  location                = azurerm_resource_group.example.location
  resource_group_name     = azurerm_resource_group.example.name
  sku_name                = "Standard"
  idle_timeout_in_minutes = 10
  zones                   = ["1"]
}
resource "azurerm_nat_gateway" "my_gateway2" {
  name                    = "nat-Gateway"
  location                = azurerm_resource_group.example.location
  resource_group_name     = azurerm_resource_group.example.name
  sku_name                = "Standard"
  idle_timeout_in_minutes = 10
  zones                   = ["1"]
}

resource "azurerm_nat_gateway_public_ip_prefix_association" "my_gateway" {
  nat_gateway_id      = azurerm_nat_gateway.my_gateway.id
  public_ip_prefix_id = azurerm_public_ip_prefix.example.id
}

resource "azurerm_nat_gateway_public_ip_association" "my_gateway" {
  nat_gateway_id       = azurerm_nat_gateway.my_gateway.id
  public_ip_address_id = azurerm_public_ip.example.id
}

resource "azurerm_nat_gateway_public_ip_prefix_association" "my_gateway2" {
  nat_gateway_id      = azurerm_nat_gateway.my_gateway2.id
  public_ip_prefix_id = azurerm_public_ip_prefix.example.id
}

resource "azurerm_nat_gateway_public_ip_association" "my_gateway2" {
  nat_gateway_id       = azurerm_nat_gateway.my_gateway2.id
  public_ip_address_id = azurerm_public_ip.example.id
}
