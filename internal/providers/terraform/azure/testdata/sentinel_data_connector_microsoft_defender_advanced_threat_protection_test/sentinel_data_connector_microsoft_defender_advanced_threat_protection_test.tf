provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_log_analytics_workspace" "sentinel_data_connector_microsoft_defender_advanced_threat_protection_with_solution" {
  name                = "example-workspace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_solution" "example" {
  solution_name         = "exampleinsights"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  workspace_resource_id = azurerm_log_analytics_workspace.sentinel_data_connector_microsoft_defender_advanced_threat_protection_with_solution.id
  workspace_name        = azurerm_log_analytics_workspace.sentinel_data_connector_microsoft_defender_advanced_threat_protection_with_solution.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/SecurityInsights"
  }
}

resource "azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection" "example" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_solution.example.workspace_resource_id
}

resource "azurerm_log_analytics_workspace" "sentinel_data_connector_microsoft_defender_advanced_threat_protection" {
  name                = "example-workspace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection" "example2" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.sentinel_data_connector_microsoft_defender_advanced_threat_protection.id
}
