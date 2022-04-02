provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

# Add example resources for SentinelDataConnectorAzureActiveDirectory below

resource "azurerm_log_analytics_workspace" "sentinel_data_connector_azure_active_directory" {
  name                = "example-workspace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_solution" "sentinel_data_connector_azure_active_directory" {
  solution_name         = "SecurityInsights"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  workspace_resource_id = azurerm_log_analytics_workspace.sentinel_data_connector_azure_active_directory.id
  workspace_name        = azurerm_log_analytics_workspace.sentinel_data_connector_azure_active_directory.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/SecurityInsights"
  }
}

resource "azurerm_sentinel_data_connector_azure_active_directory" "sentinel_data_connector_azure_active_directory" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_solution.sentinel_data_connector_azure_active_directory.workspace_resource_id
}
