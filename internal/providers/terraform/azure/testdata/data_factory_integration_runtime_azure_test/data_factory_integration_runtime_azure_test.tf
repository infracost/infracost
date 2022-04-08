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

resource "azurerm_data_factory_integration_runtime_azure" "default_with_usage" {
  name            = "runtime-default"
  data_factory_id = azurerm_data_factory.example.id
  location        = azurerm_resource_group.example.location
}

resource "azurerm_data_factory_integration_runtime_azure" "gp" {
  name            = "runtime-gp"
  data_factory_id = azurerm_data_factory.example.id
  location        = azurerm_resource_group.example.location
  compute_type    = "General"
  core_count      = 8
}

resource "azurerm_data_factory_integration_runtime_azure" "co" {
  name            = "runtime-co"
  data_factory_id = azurerm_data_factory.example.id
  location        = azurerm_resource_group.example.location
  compute_type    = "ComputeOptimized"
  core_count      = 16
}

resource "azurerm_data_factory_integration_runtime_azure" "mo" {
  name            = "runtime-mo"
  data_factory_id = azurerm_data_factory.example.id
  location        = azurerm_resource_group.example.location
  compute_type    = "MemoryOptimized"
  core_count      = 272
}
