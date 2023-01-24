provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-rg"
  location = "westus"
}

output "resource_group_name" {
  value = azurerm_resource_group.example.name
}
