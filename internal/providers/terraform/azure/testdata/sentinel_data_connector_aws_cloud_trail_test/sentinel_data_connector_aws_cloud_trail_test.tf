provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

# Add example resources for SentinelDataConnectorAwsCloudTrail below

resource "azurerm_log_analytics_workspace" "sentinel_data_connector_aws_cloud_trail" {
  name                = "example-workspace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_solution" "sentinel_data_connector_aws_cloud_trail" {
  solution_name         = "exampleinsights"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  workspace_resource_id = azurerm_log_analytics_workspace.sentinel_data_connector_aws_cloud_trail.id
  workspace_name        = azurerm_log_analytics_workspace.sentinel_data_connector_aws_cloud_trail.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/SecurityInsights"
  }
}

resource "azurerm_sentinel_data_connector_aws_cloud_trail" "sentinel_data_connector_aws_cloud_trail" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_solution.sentinel_data_connector_aws_cloud_trail.workspace_resource_id
  aws_role_arn               = "arn:aws:iam::000000000000:role/role1"
}
