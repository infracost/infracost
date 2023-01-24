provider "azurerm" {
  skip_provider_registration = true
  features {}
}

variable "resource_group_name" {}

resource "azurerm_dns_zone" "example" {
  name                = "mydomain.com"
  resource_group_name = var.resource_group_name
}
