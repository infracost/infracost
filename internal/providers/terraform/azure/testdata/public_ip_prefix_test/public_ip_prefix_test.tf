provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "eastus"
}
resource "azurerm_public_ip_prefix" "example" {
  name                = "acceptanceTestPublicIpPrefix1"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  prefix_length = 31

  tags = {
    environment = "Production"
  }
}
