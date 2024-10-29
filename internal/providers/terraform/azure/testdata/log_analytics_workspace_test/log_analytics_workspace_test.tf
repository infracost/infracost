provider "azurerm" {
  skip_provider_registration = true
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_log_analytics_workspace" "per_gb_data_ingestion" {
  name                = "acctest-01"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_workspace" "capacity_gb_data_ingestion" {
  name                               = "acctest-02"
  location                           = azurerm_resource_group.example.location
  resource_group_name                = azurerm_resource_group.example.name
  sku                                = "CapacityReservation"
  reservation_capacity_in_gb_per_day = 100
}

resource "azurerm_log_analytics_workspace" "capacity_gb_data_ingestion_without_specification" {
  name                = "acctest-03"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "CapacityReservation"
}

resource "azurerm_log_analytics_workspace" "log_data_retention_free" {
  name                = "acctest-04"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_log_analytics_workspace" "log_data_retention_with_usage" {
  name                = "acctest-05"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
  retention_in_days   = 33
}

resource "azurerm_log_analytics_workspace" "log_data_retention_without_usage" {
  name                = "acctest-08"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
  retention_in_days   = 33
}

resource "azurerm_log_analytics_workspace" "log_data_export" {
  name                = "acctest-06"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}


resource "azurerm_log_analytics_workspace" "unsupported_legacy_workspace" {
  for_each            = toset(["Unlimited", "Standard", "Premium", "PerNode"])
  name                = "acctest-unsupported-${each.key}"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = each.value
}

resource "azurerm_log_analytics_workspace" "per_gb_sentinel_data_ingestion" {
  name                = "acctest-10"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_solution" "per_gb_sentinel_solution" {
  solution_name         = "SecurityInsights"
  location              = azurerm_resource_group.example.location
  resource_group_name   = azurerm_resource_group.example.name
  workspace_resource_id = azurerm_log_analytics_workspace.per_gb_sentinel_data_ingestion.id
  workspace_name        = azurerm_log_analytics_workspace.per_gb_sentinel_data_ingestion.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/SecurityInsights"
  }
}

resource "azurerm_sentinel_data_connector_aws_cloud_trail" "sentinel_data_connector_aws_cloud_trail" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_solution.per_gb_sentinel_solution.workspace_resource_id
  aws_role_arn               = "arn:aws:iam::000000000000:role/role1"
}

resource "azurerm_log_analytics_workspace" "per_gb_sentinel_data_ingestion_with_usage" {
  name                = "acctest-10"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_sentinel_data_connector_azure_advanced_threat_protection" "example" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.per_gb_sentinel_data_ingestion_with_usage.id
}

resource "azurerm_log_analytics_workspace" "capacity_sentinel_data_ingestion" {
  name                               = "acctest-10"
  location                           = azurerm_resource_group.example.location
  resource_group_name                = azurerm_resource_group.example.name
  sku                                = "CapacityReservation"
  reservation_capacity_in_gb_per_day = 100
}

resource "azurerm_sentinel_data_connector_azure_active_directory" "example" {
  name                       = "example"
  log_analytics_workspace_id = azurerm_log_analytics_workspace.capacity_sentinel_data_ingestion.id
}

resource "azurerm_log_analytics_workspace" "per_gb_basic_data_ingestion_with_usage" {
  name                = "acctest-data-ingest"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}

resource "azurerm_log_analytics_workspace" "archive_data_with_usage" {
  name                = "acctest-archive"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
}
