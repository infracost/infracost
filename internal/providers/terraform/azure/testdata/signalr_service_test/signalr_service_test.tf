provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_signalr_service" "example" {
  for_each = toset(["Free_F1", "Standard_S1", "Premium_P1"])

  name                = "tfex-signalr"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  sku {
    name     = each.value
    capacity = 1
  }
}

resource "azurerm_signalr_service" "example_with_usage" {
  for_each = toset(["Standard_S1", "Premium_P1"])

  name                = "tfex-signalr"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  sku {
    name     = each.value
    capacity = 5
  }
}
