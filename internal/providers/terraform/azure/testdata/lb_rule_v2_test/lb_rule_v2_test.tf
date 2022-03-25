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
resource "azurerm_public_ip" "example" {
  name                = "PublicIPForLB"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name
  allocation_method   = "Static"
}

resource "azurerm_lb" "example" {
  name                = "TestLoadBalancer"
  location            = "West US"
  resource_group_name = azurerm_resource_group.example.name

  frontend_ip_configuration {
    name                 = "PublicIPAddress"
    public_ip_address_id = azurerm_public_ip.example.id
  }
}

resource "azurerm_lb_rule" "rules" {
  resource_group_name            = azurerm_resource_group.example.name
  loadbalancer_id                = azurerm_lb.example.id
  name                           = "LBRule"
  protocol                       = "Tcp"
  frontend_port                  = 3389
  backend_port                   = 3389
  frontend_ip_configuration_name = "PublicIPAddress"
}
