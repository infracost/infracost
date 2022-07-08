provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

# Add example resources for Iothub below
resource "azurerm_iothub" "example" {
  name                = "Example-IoTHub"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location

  sku {
    name     = "S1"
    capacity = 2
  }

  tags = {
    purpose = "testing"
  }
}

resource "azurerm_iothub_dps" "example" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  allocation_policy   = "Hashed"

  sku {
    name     = "S1"
    capacity = 2
  }

  tags = {
    purpose = "testing"
  }
}
