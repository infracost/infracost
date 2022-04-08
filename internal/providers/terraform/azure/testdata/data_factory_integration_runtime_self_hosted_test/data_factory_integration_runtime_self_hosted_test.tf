provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_data_factory" "example" {
  name                = "example"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_data_factory_integration_runtime_self_hosted" "example" {
  name            = "example"
  data_factory_id = azurerm_data_factory.example.id
}

resource "azurerm_data_factory_integration_runtime_self_hosted" "with_usage" {
  name            = "with-usage"
  data_factory_id = azurerm_data_factory.example.id
}
