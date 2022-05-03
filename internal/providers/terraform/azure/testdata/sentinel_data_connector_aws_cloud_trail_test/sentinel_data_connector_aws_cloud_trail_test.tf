provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_log_analytics_workspace" "sentinel_data_connector_aws_cloud_trail_with_solution" {
  name                = "example-workspace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_solution" "example" {
  solution_name         = "exampleinsights"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  workspace_resource_id = azurerm_log_analytics_workspace.sentinel_data_connector_aws_cloud_trail_with_solution.id
  workspace_name        = azurerm_log_analytics_workspace.sentinel_data_connector_aws_cloud_trail_with_solution.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/SecurityInsights"
  }
}

resource "azurerm_sentinel_data_connector_aws_cloud_trail" "example" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_solution.example.workspace_resource_id
  aws_role_arn               = "arn:aws:iam::000000000000:role/role1"
}

resource "azurerm_log_analytics_workspace" "sentinel_data_connector_aws_cloud_trail" {
  name                = "example-workspace"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_sentinel_data_connector_aws_cloud_trail" "example2" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.sentinel_data_connector_aws_cloud_trail.id
  aws_role_arn               = "arn:aws:iam::000000000000:role/role1"
}
