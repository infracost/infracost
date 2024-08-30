terraform {
  required_providers {
    azurerm-v2 = {
      source  = "hashicorp/azurerm"
      version = "~> 2.0"
    }
  }
}

provider "azurerm-v2" {
  features {}
  skip_provider_registration = true
}

provider "azurerm" {
  features {}
  skip_provider_registration = true
}
resource "azurerm_resource_group" "example" {
  name     = "LoadBalancerRG"
  location = "eastus"
}
resource "azurerm_public_ip" "example_withDefaultSku" {
  name                = "PublicIPForLB"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_lb" "example_withDefaultSku" {
  name                = "TestLoadBalancer"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = azurerm_public_ip.example_withDefaultSku.id
  }
}

resource "azurerm_lb_rule" "rules_withDefaultSku" {
  resource_group_name            = azurerm_resource_group.example.name
  loadbalancer_id                = azurerm_lb.example_withDefaultSku.id
  name                           = "LBRule"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "PublicIPAddress"
}

resource "azurerm_public_ip" "example_withBasicSku" {
  name                = "PublicIPForLB"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name
  allocation_method   = "Static"
  sku                 = "Basic"
}

resource "azurerm_lb" "example_withBasicSku" {
  name                = "TestBasicLoadBalancer"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Basic"

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = azurerm_public_ip.example_withBasicSku.id
  }
}

resource "azurerm_lb_rule" "rules_withBasicSku" {
  resource_group_name            = azurerm_resource_group.example.name
  loadbalancer_id                = azurerm_lb.example_withBasicSku.id
  name                           = "LBRule"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "PublicIPAddress"
}

resource "azurerm_public_ip" "example_withStandardSku" {
  name                = "PublicIPForLB"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_lb" "example_withStandardSku" {
  name                = "TestBasicLoadBalancer"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "Standard"

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = azurerm_public_ip.example_withStandardSku.id
  }
}

resource "azurerm_lb_rule" "rules_withStandardSku" {
  resource_group_name            = azurerm_resource_group.example.name
  loadbalancer_id                = azurerm_lb.example_withStandardSku.id
  name                           = "LBRule"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "PublicIPAddress"
}
