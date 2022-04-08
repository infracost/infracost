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

resource "azurerm_data_factory_integration_runtime_managed" "default_with_usage" {
  name            = "runtime-default"
  data_factory_id = azurerm_data_factory.example.id
  location        = azurerm_resource_group.example.location

  node_size = "Standard_D8_v3"
}

resource "azurerm_data_factory_integration_runtime_managed" "standard_ahb_v2" {
  name            = "runtime-standard"
  data_factory_id = azurerm_data_factory.example.id
  location        = azurerm_resource_group.example.location

  node_size       = "Standard_D1_v2"
  number_of_nodes = 4
  edition         = "Standard"
  license_type    = "BasePrice"
}

resource "azurerm_data_factory_integration_runtime_managed" "enterprise_with_license" {
  name            = "runtime-enterprise"
  data_factory_id = azurerm_data_factory.example.id
  location        = azurerm_resource_group.example.location

  node_size    = "Standard_E64_v3"
  edition      = "Enterprise"
  license_type = "LicenseIncluded"
}
